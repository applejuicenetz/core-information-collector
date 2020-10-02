import java.awt.*;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.image.BufferedImage;
import java.io.*;
import java.net.HttpURLConnection;
import java.net.URL;
import java.net.URLConnection;
import java.text.DecimalFormat;
import java.text.NumberFormat;
import java.util.Locale;
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

    private TrayIcon trayIcon;

    private final ImageIcon appIcon = new ImageIcon(getClass().getResource("icon.png"));

    private final DocumentBuilderFactory dbf = DocumentBuilderFactory.newInstance();

    private final XPathFactory xpf = XPathFactory.newInstance();

    private String coreVersion;
    private String coreSystem;

    private long coreCredits;
    private long coreConnections;
    private long coreSessionUpload;
    private long coreSessionDownload;
    private long coreUploadSpeed;
    private long coreDownloadSpeed;

    private long coreUploads;
    private long coreDownloads;
    private long coreDownloadsReady;

    private long networkUser;
    private long networkFiles;
    private double networkFileSize;

    private int errorCountGet = 0;

    private int errorCountPost = 0;

    public Presets presets;

    public static void main(String[] args) {
        new Collector();
    }

    public Collector() {
        presets = new Presets();

        configureSystemTray();

        try {
            getCoreVersion();
        } catch (Exception e) {
            e.printStackTrace();
            JOptionPane.showMessageDialog(null, "Konnte keine Verbindung zum Core aufbauen.", APP_NAME, JOptionPane.ERROR_MESSAGE, appIcon);
            System.exit(-1);
        }

        run();

        Timer timer = new Timer();
        timer.schedule(this, 1000, presets.getInterval());
    }

    public void getCoreVersion() throws Exception {
        String url = String.format("%s:%s/xml/information.xml?password=%s", presets.getCoreHost(), presets.getCorePort(), presets.getCorePassword());

        String payload;

        try {
            payload = getHTTPResource(url);

        } catch (Exception e) {
            System.out.println(e.toString());
            return;
        }

        DocumentBuilder db = dbf.newDocumentBuilder();
        Document document = db.parse(new InputSource(new StringReader(payload)));

        XPath xpath = xpf.newXPath();
        Element generalinformationVersion = (Element) xpath.evaluate("/applejuice/generalinformation/version", document, XPathConstants.NODE);
        Element generalinformationSystem = (Element) xpath.evaluate("/applejuice/generalinformation/system", document, XPathConstants.NODE);

        coreVersion = generalinformationVersion.getFirstChild().getNodeValue();
        coreSystem = generalinformationSystem.getFirstChild().getNodeValue();
    }

    public void run() {
        getAppleJuiceInformation();

        updateInfoLine();

        if (!presets.getForwardUrl().equals("off")) {
            try {
                forward();
                errorCountPost = 0;
            } catch (Exception e) {
                e.printStackTrace();
                errorCountPost++;
            }

            if (errorCountPost > 30) {
                System.err.println("to many errors");
                JOptionPane.showMessageDialog(null, "Konnte die API URL zu oft nicht erreichen, keine Internetverbindung?", APP_NAME, JOptionPane.ERROR_MESSAGE, appIcon);
                System.exit(-1);
            }
        }
    }

    private void getAppleJuiceInformation() {
        String payload;

        String url = String.format("%s:%s/xml/modified.xml?password=%s", presets.getCoreHost(), presets.getCorePort(), presets.getCorePassword());

        if (errorCountGet > 10) {
            System.err.println("to many errors");
            JOptionPane.showMessageDialog(null, "Core konnte zu oft nicht erreicht werden, Daten korrekt?", APP_NAME, JOptionPane.ERROR_MESSAGE, appIcon);
            System.exit(-1);
        }

        try {
            payload = getHTTPResource(url);
            errorCountGet = 0;
        } catch (Exception e) {
            System.out.println(e.toString());
            errorCountGet++;
            return;
        }

        try {
            handleAppleJuiceInformation(payload);
        } catch (Exception e) {
            e.printStackTrace();
        }
    }

    private void handleAppleJuiceInformation(String payload) throws Exception {
        DocumentBuilder db = dbf.newDocumentBuilder();
        Document document = db.parse(new InputSource(new StringReader(payload)));

        XPath xpath = xpf.newXPath();
        Element coreInfo = (Element) xpath.evaluate("/applejuice/information", document, XPathConstants.NODE);
        Element networkInfo = (Element) xpath.evaluate("/applejuice/networkinfo", document, XPathConstants.NODE);

        double coreDownloadsReadyCount = (double) xpath.evaluate("count(/applejuice/download[@ready=0])", document, XPathConstants.NUMBER);
        coreDownloadsReady = Double.valueOf(coreDownloadsReadyCount).longValue();

        coreUploads = document.getElementsByTagName("upload").getLength();
        coreDownloads = document.getElementsByTagName("download").getLength();

        coreCredits = Long.parseLong(coreInfo.getAttribute("credits"));
        coreConnections = Long.parseLong(coreInfo.getAttribute("openconnections"));
        coreSessionUpload = Long.parseLong(coreInfo.getAttribute("sessionupload"));
        coreSessionDownload = Long.parseLong(coreInfo.getAttribute("sessiondownload"));
        coreUploadSpeed = Long.parseLong(coreInfo.getAttribute("uploadspeed"));
        coreDownloadSpeed = Long.parseLong(coreInfo.getAttribute("downloadspeed"));

        networkUser = Long.parseLong(networkInfo.getAttribute("users"));
        networkFiles = Long.parseLong(networkInfo.getAttribute("files"));

        networkFileSize = Double.parseDouble(networkInfo.getAttribute("filesize"));
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

    public void forward() throws Exception {
        String charset = "UTF-8";
        URLConnection connection = new URL(presets.getForwardUrl()).openConnection();
        connection.setDoOutput(true); // Triggers POST.
        connection.setRequestProperty("Accept-Charset", charset);
        if (!presets.getForwardToken().isEmpty()) {
            connection.setRequestProperty("Authorization", "Token " + presets.getForwardToken());

        }
        connection.setRequestProperty("Content-Type", "text/plain;charset=" + charset);

        try (OutputStream output = connection.getOutputStream()) {
            output.write(getFormattedForwardLine().getBytes(charset));
        }

        InputStream response = connection.getInputStream();
    }

    private String getFormattedForwardLine() {
        return presets.getForwardLine()
                .replace("%coreVersion%", coreVersion)
                .replace("%coreSystem%", coreSystem)
                .replace("%coreCredits%", readableFileSize(coreCredits))
                .replace("%coreConnections%", readableFileSize(coreConnections))
                .replace("%coreSessionUpload%", readableFileSize(coreSessionUpload))
                .replace("%coreSessionDownload%", readableFileSize(coreSessionDownload))
                .replace("%coreUploadSpeed%", readableFileSize(coreUploadSpeed) + "/s")
                .replace("%coreDownloadSpeed%", readableFileSize(coreDownloadSpeed) + "/s")

                .replace("%coreUploads%", Long.toString(coreUploads))
                .replace("%coreDownloads%", Long.toString(coreDownloads))
                .replace("%coreDownloadsReady%", Long.toString(coreDownloadsReady));
    }

    private void updateInfoLine() {
        String Info = presets.getInfoLine()
                .replace("%coreVersion%", coreVersion)
                .replace("%coreSystem%", coreSystem)
                .replace("%coreCredits%", readableFileSize(coreCredits))
                .replace("%coreConnections%", readableFileSize(coreConnections))
                .replace("%coreSessionUpload%", readableFileSize(coreSessionUpload))
                .replace("%coreSessionDownload%", readableFileSize(coreSessionDownload))
                .replace("%coreUploadSpeed%", readableFileSize(coreUploadSpeed) + "/s")
                .replace("%coreDownloadSpeed%", readableFileSize(coreDownloadSpeed) + "/s")

                .replace("%coreUploads%", Long.toString(coreUploads))
                .replace("%coreDownloads%", Long.toString(coreDownloads))
                .replace("%coreDownloadsReady%", Long.toString(coreDownloadsReady))

                .replace("%networkUser%", Long.toString(networkUser))
                .replace("%networkFiles%", NumberFormat.getNumberInstance(Locale.GERMAN).format(networkFiles))
                .replace("%networkFileSize%", readableNetworkShareSize(networkFileSize, 0));

        System.out.println(Info);

        if (SystemTray.isSupported()) {
            trayIcon.setToolTip(Info);
        }
    }

    private void configureSystemTray() {
        if (SystemTray.isSupported()) {
            SystemTray systemTray = SystemTray.getSystemTray();

            PopupMenu menu = this.createMenu();

            try {
                BufferedImage trayIconImage = ImageIO.read(getClass().getResource("icon.png"));
                int trayIconWidth = new TrayIcon(trayIconImage).getSize().width;
                trayIcon = new TrayIcon(trayIconImage.getScaledInstance(trayIconWidth, -1, Image.SCALE_SMOOTH), APP_NAME, menu);

                systemTray.add(trayIcon);
            } catch (Exception e) {
                e.printStackTrace();
            }
        }
    }

    public void actionPerformed(ActionEvent ev) {
        switch (ev.getActionCommand()) {
            case "about":
                JOptionPane.showMessageDialog(null, "Version " + getVersion(), APP_NAME, JOptionPane.INFORMATION_MESSAGE, appIcon);
                break;

            case "quit":
                System.exit(0);
                break;
        }
    }

    public static String readableFileSize(long size) {
        if (size <= 0) return "0";
        final String[] units = new String[]{"B", "kB", "MB", "GB", "TB"};
        int digitGroups = (int) (Math.log10(size) / Math.log10(1024));
        return new DecimalFormat("#,##0.##").format(size / Math.pow(1024, digitGroups)) + units[digitGroups];
    }

    public String readableNetworkShareSize(double share, long faktor) {

        if (share == 0) {
            return "0,00 MB";
        }

        if (faktor == 0) { // selbst entscheiden
            if (share / 1024 < 1024) {
                faktor = 1024;
            } else if (share / 1048576 < 1024) {
                faktor = 1048576;
            } else {
                faktor = 1;
            }
        }

        share = share / faktor;
        String result = Double.toString(share);

        if (result.indexOf(".") + 3 < result.length()) {
            result = result.substring(0, result.indexOf(".") + 3);
        }

        result = result.replace('.', ',');
        if (faktor == 1) {
            result += "MB";
        } else if (faktor == 1024) {
            result += "GB";
        } else if (faktor == 1048576) {
            result += "TB";
        } else {
            result += "??";
        }

        return result;
    }

    private PopupMenu createMenu() {
        PopupMenu menu = new PopupMenu(APP_NAME + getVersion());

        MenuItem menuAbout = new MenuItem("About");
        menuAbout.setActionCommand("about");
        menu.add(menuAbout);

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
            return "SNAPSHOT";
        }
    }
}
