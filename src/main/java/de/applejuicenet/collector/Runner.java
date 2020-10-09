package de.applejuicenet.collector;

import org.tinylog.Logger;
import org.w3c.dom.Document;
import org.w3c.dom.Element;
import org.xml.sax.InputSource;

import javax.swing.*;
import javax.swing.table.DefaultTableModel;
import javax.xml.parsers.DocumentBuilder;
import javax.xml.parsers.DocumentBuilderFactory;
import javax.xml.xpath.XPath;
import javax.xml.xpath.XPathConstants;
import javax.xml.xpath.XPathFactory;
import java.awt.*;
import java.io.InputStream;
import java.io.OutputStream;
import java.io.StringReader;
import java.net.URL;
import java.net.URLConnection;
import java.text.NumberFormat;
import java.util.Timer;
import java.util.*;

public class Runner extends TimerTask {

    public final String APP_NAME = "appleJuice Core Information Collector";

    public final ImageIcon appIcon = new ImageIcon(getClass().getResource("/resources/icon.png"));

    private final DocumentBuilderFactory dbf = DocumentBuilderFactory.newInstance();

    private final XPathFactory xpf = XPathFactory.newInstance();

    private final TreeMap<String, String> replacer = new TreeMap<>();

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

    private final Config config;

    private final Tray tray;

    public static void main(String[] args) {
        new Runner();
    }

    public Runner() {
        config = new Config();

        tray = new Tray(this);

        Version version = new Version(this);

        version.check4update();

        run();

        Timer timer = new Timer();
        timer.schedule(this, 1000, config.getInterval());
    }

    public void run() {
        try {
            prepareCoreVersion();
        } catch (Exception e) {
            Logger.error(e);
            return;
        }

        try {
            prepareAppleJuiceInformation();
        } catch (Exception e) {
            Logger.error(e);
            return;
        }

        updateReplacer();

        String InfoLine = replacePlaceHolder(config.getInfoLine());

        Logger.info(InfoLine);

        if (SystemTray.isSupported()) {
            tray.setToolTip(InfoLine);
        }

        if (!config.getForwardUrl().equals("off")) {
            try {
                forward();
            } catch (Exception e) {
                Logger.error(e);
            }
        }
    }

    private void prepareCoreVersion() throws Exception {
        String url = String.format("%s:%s/xml/information.xml?password=%s", config.getCoreHost(), config.getCorePort(), config.getCorePassword());

        String payload;

        payload = Http.get(url);

        handleAppleCoreVersion(payload);
    }

    private void handleAppleCoreVersion(String payload) throws Exception {
        DocumentBuilder db = dbf.newDocumentBuilder();
        Document document = db.parse(new InputSource(new StringReader(payload)));

        XPath xpath = xpf.newXPath();
        Element generalinformationVersion = (Element) xpath.evaluate("/applejuice/generalinformation/version", document, XPathConstants.NODE);
        Element generalinformationSystem = (Element) xpath.evaluate("/applejuice/generalinformation/system", document, XPathConstants.NODE);

        coreVersion = generalinformationVersion.getFirstChild().getNodeValue();
        coreSystem = generalinformationSystem.getFirstChild().getNodeValue();
    }

    private void prepareAppleJuiceInformation() throws Exception {
        String payload;

        String url = String.format("%s:%s/xml/modified.xml?password=%s", config.getCoreHost(), config.getCorePort(), config.getCorePassword());

        payload = Http.get(url);

        handleAppleJuiceInformation(payload);
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

    private void forward() throws Exception {
        String charset = "UTF-8";
        URLConnection connection = new URL(config.getForwardUrl()).openConnection();
        connection.setDoOutput(true); // Triggers POST.
        connection.setRequestProperty("Accept-Charset", charset);
        if (!config.getForwardToken().isEmpty()) {
            connection.setRequestProperty("Authorization", "Token " + config.getForwardToken());

        }
        connection.setRequestProperty("Content-Type", "text/plain;charset=" + charset);

        try (OutputStream output = connection.getOutputStream()) {
            String Line = replacePlaceHolder(config.getForwardLine());
            output.write(Line.getBytes(charset));
        }

        InputStream response = connection.getInputStream();
    }

    private String replacePlaceHolder(String Line) {
        Iterator<Map.Entry<String, String>> it = replacer.entrySet().iterator();
        while (it.hasNext()) {
            Map.Entry<String, String> pair = it.next();
            Line = Line.replace(pair.getKey(), pair.getValue());
        }

        return Line;
    }

    private synchronized void updateReplacer() {
        replacer.put("%coreSystem%", coreSystem);
        replacer.put("%coreVersion%", coreVersion);
        replacer.put("%coreCredits%", NumberFormatter.readableFileSize(coreCredits));
        replacer.put("%coreConnections%", Long.toString(coreConnections));
        replacer.put("%coreSessionUpload%", NumberFormatter.readableFileSize(coreSessionUpload));
        replacer.put("%coreSessionDownload%", NumberFormatter.readableFileSize(coreSessionDownload));
        replacer.put("%coreUploadSpeed%", NumberFormatter.readableFileSize(coreUploadSpeed) + "/s");
        replacer.put("%coreDownloadSpeed%", NumberFormatter.readableFileSize(coreDownloadSpeed) + "/s");

        replacer.put("%coreUploads%", Long.toString(coreUploads));
        replacer.put("%coreDownloads%", Long.toString(coreDownloads));
        replacer.put("%coreDownloadsReady%", Long.toString(coreDownloadsReady));

        replacer.put("%networkUser%", Long.toString(networkUser));
        replacer.put("%networkFiles%", NumberFormat.getNumberInstance(Locale.GERMAN).format(networkFiles));
        replacer.put("%networkFileSize%", NumberFormatter.readableNetworkShareSize(networkFileSize, 0));
    }

    public void openStatusFrame() {
        DefaultTableModel model = new DefaultTableModel(
                new Object[]{"Key", "Value"}, 0
        );

        Iterator<Map.Entry<String, String>> it = replacer.entrySet().iterator();
        while (it.hasNext()) {
            Map.Entry<String, String> pair = it.next();
            model.addRow(new Object[]{pair.getKey(), pair.getValue()});
            it.remove();
        }

        JTable table = new JTable(model);
        table.setEnabled(false);

        table.getTableHeader().setOpaque(false);

        JScrollPane scrollPane = new JScrollPane(table);
        scrollPane.setPreferredSize(new Dimension(300, 250));

        JOptionPane.showMessageDialog(null, scrollPane, APP_NAME, JOptionPane.INFORMATION_MESSAGE, appIcon);
    }
}
