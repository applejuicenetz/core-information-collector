package de.applejuicenet.collector;

import dev.hivens.libtray.Tray;
import dev.hivens.libtray.TrayBuilder;
import dev.hivens.libtray.TrayEvent;
import dev.hivens.libtray.TrayMenu;
import dev.hivens.libtray.TrayMenuItem;
import org.tinylog.Logger;
import kotlin.Unit;

import javax.imageio.ImageIO;
import javax.swing.*;
import java.awt.*;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.image.BufferedImage;
import java.io.File;
import java.io.IOException;
import java.io.InputStream;
import java.util.List;
import java.util.Locale;
import java.util.Objects;

public class TaskbarAndTray implements ActionListener {

    private TrayIcon trayIcon;
    private JPopupMenu trayMenu;
    private Tray nativeTray;

    private final Runner runner;

    public TaskbarAndTray(Runner runner) {
        this.runner = runner;

        if (!runner.config.isTrayIcon() && !runner.config.isTaskBarIcon()) {
            return;
        }

        if (GraphicsEnvironment.isHeadless()) {
            return;
        }

        if (Desktop.isDesktopSupported() && Desktop.getDesktop().isSupported(Desktop.Action.APP_ABOUT)) {
            Desktop.getDesktop().setAboutHandler(e -> showAboutDialog());
        }

        if (runner.config.isTaskBarIcon() && Taskbar.isTaskbarSupported()) {
            final Taskbar taskbar = Taskbar.getTaskbar();

            if (taskbar.isSupported(Taskbar.Feature.MENU)) {
                PopupMenu taskbarMenu = createMenu();
                taskbar.setMenu(taskbarMenu);
            }
        }

        boolean nativeTrayCreated = runner.config.isTrayIcon() && isLinux() && createNativeTray();

        if (runner.config.isTrayIcon() && !nativeTrayCreated && SystemTray.isSupported()) {
            SystemTray systemTray = SystemTray.getSystemTray();

            try {
                boolean macOS = isMacOS();
                trayMenu = macOS ? null : createTrayMenu();

                BufferedImage trayIconImage = ImageIO.read(getClass().getResource("/resources/icon.png"));
                int trayIconWidth = new TrayIcon(trayIconImage).getSize().width;

                trayIcon = macOS
                        ? new TrayIcon(trayIconImage.getScaledInstance(trayIconWidth, -1, Image.SCALE_SMOOTH), Runner.APP_NAME, createMenu())
                        : new TrayIcon(trayIconImage.getScaledInstance(trayIconWidth, -1, Image.SCALE_SMOOTH), Runner.APP_NAME);
                trayIcon.addMouseListener(new java.awt.event.MouseAdapter() {

                    @Override
                    public void mouseClicked(java.awt.event.MouseEvent evt) {
                        if (SwingUtilities.isLeftMouseButton(evt) && !evt.isPopupTrigger()) {
                            runner.toggleStatusFrame();
                        }
                    }

                    @Override
                    public void mousePressed(java.awt.event.MouseEvent evt) {
                        showTrayMenu(evt);
                    }

                    @Override
                    public void mouseReleased(java.awt.event.MouseEvent evt) {
                        showTrayMenu(evt);
                    }

                });

                systemTray.add(trayIcon);
            } catch (Exception e) {
                Logger.error(e);
            }
        }
    }

    public void setToolTip(String tooltip) {
        if (nativeTray != null) {
            nativeTray.setTooltip(tooltip);
        } else if (runner.config.isTrayIcon() && trayIcon != null) {
            trayIcon.setToolTip(tooltip);
        }
    }

    public void actionPerformed(ActionEvent ev) {
        switch (ev.getActionCommand()) {
            case "status":
                runner.openStatusFrame();
                break;

            case "run":
                runner.run();
                break;

            case "config":
                File configFile = Config.getConfigFile();
                try {
                    Desktop.getDesktop().open(configFile);
                } catch (IOException e) {
                    Logger.error(e);
                }
                break;

            case "about":
                showAboutDialog();
                break;

            case "quit":
                if (nativeTray != null) {
                    nativeTray.close();
                }
                System.exit(0);
                break;
        }
    }

    private PopupMenu createMenu() {
        PopupMenu menu = new PopupMenu();

        menu.add(createMenuItem("Status", "status"));
        menu.add(createMenuItem("Execute", "run"));
        menu.add(createMenuItem("Config", "config"));
        menu.add(createMenuItem("About", "about"));
        menu.add(createMenuItem("Quit", "quit"));

        return menu;
    }

    private MenuItem createMenuItem(String label, String command) {
        MenuItem item = new MenuItem(label);
        item.setActionCommand(command);
        item.addActionListener(this);
        return item;
    }

    private JPopupMenu createTrayMenu() {
        JPopupMenu menu = new JPopupMenu();

        menu.add(createTrayMenuItem("Status", "status"));
        menu.add(createTrayMenuItem("Execute", "run"));
        menu.add(createTrayMenuItem("Config", "config"));
        menu.add(createTrayMenuItem("About", "about"));
        menu.add(createTrayMenuItem("Quit", "quit"));

        return menu;
    }

    private JMenuItem createTrayMenuItem(String label, String command) {
        JMenuItem item = new JMenuItem(label);
        item.setActionCommand(command);
        item.addActionListener(this);
        return item;
    }

    private boolean createNativeTray() {
        try (InputStream icon = Objects.requireNonNull(getClass().getResourceAsStream("/resources/icon.png"))) {
            List<TrayMenuItem> items = List.of(
                    new TrayMenuItem.Standard("status", "Status", true),
                    new TrayMenuItem.Standard("run", "Execute", true),
                    new TrayMenuItem.Standard("config", "Config", true),
                    new TrayMenuItem.Standard("about", "About", true),
                    new TrayMenuItem.Standard("quit", "Quit", true)
            );
            TrayBuilder builder = new TrayBuilder(
                    Runner.APP_NAME,
                    icon.readAllBytes(),
                    Runner.APP_NAME,
                    new TrayMenu(items),
                    null,
                    "io.github.applejuicenetz.collector.StatusNotifierItem"
            );
            nativeTray = Tray.Companion.create(builder);
            if (nativeTray == null) {
                return false;
            }
            nativeTray.onEvent(event -> {
                if (event instanceof TrayEvent.MenuItemSelected selected) {
                    actionPerformed(new ActionEvent(nativeTray, ActionEvent.ACTION_PERFORMED, selected.getId()));
                } else if (event instanceof TrayEvent.Activated) {
                    runner.toggleStatusFrame();
                }
                return Unit.INSTANCE;
            });
            Logger.info("Initialized native Linux tray using StatusNotifierItem");
            return true;
        } catch (Exception | LinkageError e) {
            Logger.error(e, "Could not initialize the native Linux tray; falling back to AWT");
            nativeTray = null;
            return false;
        }
    }

    private boolean isLinux() {
        return System.getProperty("os.name", "").toLowerCase(Locale.ROOT).contains("linux");
    }

    private boolean isMacOS() {
        return System.getProperty("os.name", "").toLowerCase(Locale.ROOT).contains("mac");
    }

    private void showTrayMenu(java.awt.event.MouseEvent event) {
        if (trayMenu == null || (!event.isPopupTrigger() && !SwingUtilities.isRightMouseButton(event))) {
            return;
        }

        SwingUtilities.invokeLater(() -> {
            if (trayMenu.isVisible()) {
                return;
            }

            trayMenu.setLocation(event.getXOnScreen(), event.getYOnScreen());
            trayMenu.setInvoker(trayMenu);
            trayMenu.setVisible(true);
        });
    }

    private void showAboutDialog() {
        SwingUtilities.invokeLater(() -> JOptionPane.showMessageDialog(null, "Version " + Version.getVersion(), Runner.APP_NAME, JOptionPane.INFORMATION_MESSAGE, Runner.appIcon));
    }
}
