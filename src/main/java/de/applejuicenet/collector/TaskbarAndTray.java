package de.applejuicenet.collector;

import org.tinylog.Logger;

import javax.imageio.ImageIO;
import javax.swing.*;
import java.awt.*;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.image.BufferedImage;
import java.io.File;
import java.io.IOException;

public class TaskbarAndTray implements ActionListener {

    private TrayIcon trayIcon;

    private final Runner runner;

    public TaskbarAndTray(Runner runner) {
        this.runner = runner;

        if (!runner.config.isTrayIcon() && !runner.config.isTaskBarIcon()) {
            return;
        }

        if (!SystemTray.isSupported() && !Taskbar.isTaskbarSupported()) {
            return;
        }

        PopupMenu menu = createMenu();

        menu.addActionListener(this);

        if (runner.config.isTaskBarIcon() && Taskbar.isTaskbarSupported()) {
            final Taskbar taskbar = Taskbar.getTaskbar();

            if (taskbar.isSupported(Taskbar.Feature.ICON_IMAGE)) {
                taskbar.setIconImage(Runner.appLogo.getImage());
            }

            if (taskbar.isSupported(Taskbar.Feature.MENU)) {
                taskbar.setMenu(menu);
            }
        }

        if (runner.config.isTrayIcon() && SystemTray.isSupported()) {
            SystemTray systemTray = SystemTray.getSystemTray();

            try {
                BufferedImage trayIconImage = ImageIO.read(getClass().getResource("/resources/icon.png"));
                int trayIconWidth = new TrayIcon(trayIconImage).getSize().width;
                trayIcon = new TrayIcon(trayIconImage.getScaledInstance(trayIconWidth, -1, Image.SCALE_SMOOTH), runner.APP_NAME, menu);

                trayIcon.addMouseListener(new java.awt.event.MouseAdapter() {

                    @Override
                    public void mouseClicked(java.awt.event.MouseEvent evt) {
                        if (evt.getClickCount() == 2) {
                            runner.openStatusFrame();
                        }
                    }

                });

                systemTray.add(trayIcon);
            } catch (Exception e) {
                Logger.error(e);
            }
        }
    }

    public void setToolTip(String tooltip) {
        if (runner.config.isTrayIcon() && trayIcon != null) {
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
                    e.printStackTrace();
                }
                break;

            case "about":
                JOptionPane.showMessageDialog(null, "Version " + Version.getVersion(), runner.APP_NAME, JOptionPane.INFORMATION_MESSAGE, runner.appIcon);
                break;

            case "quit":
                System.exit(0);
                break;
        }
    }

    private PopupMenu createMenu() {
        PopupMenu menu = new PopupMenu();

        MenuItem statusItem = new MenuItem("Status");
        statusItem.setActionCommand("status");
        menu.add(statusItem);

        MenuItem runItem = new MenuItem("Execute");
        runItem.setActionCommand("run");
        menu.add(runItem);

        MenuItem configItem = new MenuItem("Config");
        configItem.setActionCommand("config");
        menu.add(configItem);

        MenuItem menuAbout = new MenuItem("About");
        menuAbout.setActionCommand("about");
        menu.add(menuAbout);

        MenuItem menuExit = new MenuItem("Quit");
        menuExit.setActionCommand("quit");
        menu.add(menuExit);

        return menu;
    }
}
