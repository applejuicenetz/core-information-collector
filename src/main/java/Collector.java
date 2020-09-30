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
import org.w3c.dom.ls.DOMImplementationLS;
import org.w3c.dom.ls.LSSerializer;
import org.xml.sax.InputSource;

public class Collector extends TimerTask implements ActionListener {
    private static final String APP_NAME = "appleJuice Core Information Collector";

    private TrayIcon trayIcon;

    private long coreCredits;
    private long coreSessionUpload;
    private long coreSessionDownload;
    private long coreUploadSpeed;
    private long coreDownloadSpeed;

    private long networkUser;
    private long networkFiles;
    private long networkFileSize;

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
            errorCountGet = 0;

        } catch (Exception e) {
            System.out.println(e.toString());
            errorCountGet++;
            return;
        }

        try {
            payload = handleAppleJuiceInformation(payload);
        } catch (Exception e) {
            e.printStackTrace();
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
        String Info = presets.getInfoLine();

        Info = Info.replace("%coreCredits%", readableFileSize(coreCredits));
        Info = Info.replace("%coreSessionUpload%", readableFileSize(coreSessionUpload));
        Info = Info.replace("%coreSessionDownload%", readableFileSize(coreSessionDownload));
        Info = Info.replace("%coreUploadSpeed%", readableFileSize(coreUploadSpeed) + "/s");
        Info = Info.replace("%coreDownloadSpeed%", readableFileSize(coreDownloadSpeed) + "/s");

        Info = Info.replace("%networkUser%", readableFileSize(networkUser));
        Info = Info.replace("%networkFiles%", readableFileSize(networkFiles));
        Info = Info.replace("%networkFileSize%", readableFileSize(networkFileSize));

        System.out.println(Info);

        trayIcon.setToolTip(Info);
    }

    private String handleAppleJuiceInformation(String payload) throws Exception {
        DocumentBuilderFactory dbf = DocumentBuilderFactory.newInstance();
        DocumentBuilder db = dbf.newDocumentBuilder();
        Document document = db.parse(new InputSource(new StringReader(payload)));

        XPathFactory xpf = XPathFactory.newInstance();
        XPath xpath = xpf.newXPath();
        Element coreInfo = (Element) xpath.evaluate("/applejuice/information", document, XPathConstants.NODE);
        Element networkInfo = (Element) xpath.evaluate("/applejuice/networkinfo", document, XPathConstants.NODE);

        coreCredits = Long.parseLong(coreInfo.getAttribute("credits"));
        coreSessionUpload = Long.parseLong(coreInfo.getAttribute("sessionupload"));
        coreSessionDownload = Long.parseLong(coreInfo.getAttribute("sessiondownload"));
        coreUploadSpeed = Long.parseLong(coreInfo.getAttribute("uploadspeed"));
        coreDownloadSpeed = Long.parseLong(coreInfo.getAttribute("downloadspeed"));

        networkUser = Long.parseLong(networkInfo.getAttribute("users"));
        networkFiles = Long.parseLong(networkInfo.getAttribute("files"));
        networkFileSize = (long) Float.parseFloat(networkInfo.getAttribute("filesize"));

        DOMImplementationLS lsImpl = (DOMImplementationLS) coreInfo.getOwnerDocument().getImplementation().getFeature("LS", "3.0");
        LSSerializer serializer = lsImpl.createLSSerializer();
        serializer.getDomConfig().setParameter("xml-declaration", false); //by default its true, so set it to false to get String without xml-declaration
        payload = serializer.writeToString(coreInfo);

        return payload;
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
