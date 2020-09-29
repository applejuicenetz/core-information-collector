import java.awt.*;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.image.BufferedImage;
import java.io.*;
import java.net.HttpURLConnection;
import java.net.URL;
import java.net.URLConnection;
import java.text.DecimalFormat;
import java.util.Timer;
import java.util.TimerTask;

import javax.imageio.ImageIO;
import javax.swing.*;
import javax.xml.parsers.DocumentBuilder;
import javax.xml.parsers.DocumentBuilderFactory;
import javax.xml.xpath.XPath;
import javax.xml.xpath.XPathConstants;
import javax.xml.xpath.XPathFactory;

import org.w3c.dom.Document;
import org.w3c.dom.Element;
import org.xml.sax.InputSource;

public class Collector extends TimerTask implements ActionListener {
    private static final String APP_NAME = "appleJuice Core Information Collector";
    private static final String INFO_LINE = "Credits %s - Sessionupload %s - sessiondownload %s - uploadspeed %s/s - downloadspeed %s/s";

    private TrayIcon trayIcon;

    private long credits;
    private long sessionupload;
    private long sessiondownload;
    private long uploadspeed;
    private long downloadspeed;

    private int errorCountGet = 0;

    private int errorCountPost = 0;

    public Presets presets;

    public static void main(String[] args) {
        if (SystemTray.isSupported()) {
            new Collector();
        } else {
            JOptionPane.showMessageDialog(null, "SystemTray not supported", APP_NAME, JOptionPane.ERROR_MESSAGE);
        }
    }

    public Collector() {
        SystemTray systemTray = SystemTray.getSystemTray();

        presets = new Presets();

        PopupMenu menu = this.createMenu();

        try {
            BufferedImage trayIconImage = ImageIO.read(getClass().getResource("icon.png"));
            int trayIconWidth = new TrayIcon(trayIconImage).getSize().width;
            trayIcon = new TrayIcon(trayIconImage.getScaledInstance(trayIconWidth, -1, Image.SCALE_SMOOTH), "Collector", menu);

            systemTray.add(trayIcon);
        } catch (Exception ex) {
            ex.printStackTrace();
        }

        run();

        Timer timer = new Timer();
        timer.schedule(this, 1000, presets.getInterval());
    }

    public void run() {
        String payload;

        String url = String.format("%s:%s/xml/modified.xml?filter=informations&password=%s", presets.getCoreHost(), presets.getCorePort(), presets.getCorePassword());

        try {
            payload = getHTTPResource(url);

        } catch (Exception e) {
            System.out.println(e.toString());
            return;
        }

        try {
            printAppleJuiceInformation(payload);
            errorCountGet = 0;
        } catch (Exception e) {
            e.printStackTrace();
            errorCountGet++;
        }

        updateTooltip();

        if (errorCountGet > 10) {
            System.err.println("to many errors");
            JOptionPane.showMessageDialog(null, "Core konnte zu oft nicht erreicht werden, Daten korrekt?", APP_NAME, JOptionPane.ERROR_MESSAGE);
            System.exit(-1);
        }

        if ("" != presets.getCollector()) {
            try {
                post(payload);
                errorCountPost = 0;
            } catch (Exception e) {
                e.printStackTrace();
                errorCountPost++;
            }

            if (errorCountPost > 60) {
                System.err.println("to many errors");
                JOptionPane.showMessageDialog(null, "Konnte den Discord Bot zu oft nicht erreichen, keine Internetverbindung?", APP_NAME, JOptionPane.ERROR_MESSAGE);
                System.exit(-1);
            }
        }

    }

    private void updateTooltip() {
        String Info = String.format(
                INFO_LINE,
                readableFileSize(credits),
                readableFileSize(sessionupload),
                readableFileSize(sessiondownload),
                readableFileSize(uploadspeed),
                readableFileSize(downloadspeed)
        );

        System.out.println(Info);

        trayIcon.setToolTip(Info);
    }

    private void printAppleJuiceInformation(String payload) throws Exception {
        DocumentBuilderFactory dbf = DocumentBuilderFactory.newInstance();
        DocumentBuilder db = dbf.newDocumentBuilder();
        Document document = db.parse(new InputSource(new StringReader(payload)));

        XPathFactory xpf = XPathFactory.newInstance();
        XPath xpath = xpf.newXPath();
        Element ajInfos = (Element) xpath.evaluate("/applejuice/information", document, XPathConstants.NODE);

        credits = Long.parseLong(ajInfos.getAttribute("credits"));
        sessionupload = Long.parseLong(ajInfos.getAttribute("sessionupload"));
        sessiondownload = Long.parseLong(ajInfos.getAttribute("sessiondownload"));
        uploadspeed = Long.parseLong(ajInfos.getAttribute("uploadspeed"));
        downloadspeed = Long.parseLong(ajInfos.getAttribute("downloadspeed"));
    }

    public static String getHTTPResource(String urlToRead) throws Exception {
        StringBuilder result = new StringBuilder();
        URL url = new URL(urlToRead);
        HttpURLConnection conn = (HttpURLConnection) url.openConnection();
        conn.setRequestMethod("GET");
        BufferedReader rd = new BufferedReader(new InputStreamReader(conn.getInputStream()));
        String line;
        while ((line = rd.readLine()) != null) {
            result.append(line);
        }
        rd.close();
        return result.toString();
    }

    public void post(String payload) throws Exception {
        String charset = "UTF-8";
        URLConnection connection = new URL(presets.getCollector()).openConnection();
        connection.setDoOutput(true); // Triggers POST.
        connection.setRequestProperty("Accept-Charset", charset);
        connection.setRequestProperty("Authorization", "Token " + presets.getCollectorToken());
        connection.setRequestProperty("Content-Type", "application/xml;charset=" + charset);

        try (OutputStream output = connection.getOutputStream()) {
            output.write(payload.getBytes(charset));
        }

        InputStream response = connection.getInputStream();

        System.out.println(response.toString());
    }

    public void actionPerformed(ActionEvent ev) {
        if (ev.getActionCommand().equals("quit")) {
            System.exit(0);
        }
    }

    public static String readableFileSize(long size) {
        if (size <= 0) return "0";
        final String[] units = new String[]{"B", "kB", "MB", "GB", "TB"};
        int digitGroups = (int) (Math.log10(size) / Math.log10(1024));
        return new DecimalFormat("#,##0.#").format(size / Math.pow(1024, digitGroups)) + units[digitGroups];
    }

    private PopupMenu createMenu() {
        PopupMenu menu = new PopupMenu("appleJuice Core Collector " + getVersion());

        MenuItem menuExit = new MenuItem("Quit");
        menuExit.setActionCommand("quit");
        menu.add(menuExit);

        menu.addActionListener(this);

        return menu;
    }

    public static String getVersion() {
        try {
            return Collector.class.getPackage().getImplementationVersion();
        } catch (Exception e) {
            return "TEST";
        }
    }
}
