package de.applejuicenet.collector;

import org.tinylog.Logger;

import javax.imageio.ImageIO;
import javax.swing.*;
import java.awt.*;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.image.BufferedImage;

public class Tray implements ActionListener {

    private TrayIcon trayIcon;

    private final Runner runner;

    public Tray(Runner runner) {
        this.runner = runner;

        if (SystemTray.isSupported()) {
            SystemTray systemTray = SystemTray.getSystemTray();

            PopupMenu menu = createMenu();

            menu.addActionListener(this);

            try {
                BufferedImage trayIconImage = ImageIO.read(getClass().getResource("/resources/icon.png"));
                int trayIconWidth = new TrayIcon(trayIconImage).getSize().width;
                trayIcon = new TrayIcon(trayIconImage.getScaledInstance(trayIconWidth, -1, Image.SCALE_SMOOTH), runner.APP_NAME, menu);

                systemTray.add(trayIcon);
            } catch (Exception e) {
                Logger.error(e);
            }
        }
    }

    public void setToolTip(String tooltip) {
        trayIcon.setToolTip(tooltip);
    }


    public void actionPerformed(ActionEvent ev) {
        switch (ev.getActionCommand()) {
            case "status":
                runner.openStatusFrame();
                break;

            case "config":
                runner.openConfigFrame();
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
