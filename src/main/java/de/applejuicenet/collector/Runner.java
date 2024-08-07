package de.applejuicenet.collector;

import com.eclipsesource.json.Json;
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
import java.awt.event.KeyAdapter;
import java.awt.event.KeyEvent;
import java.io.InputStream;
import java.io.OutputStream;
import java.io.StringReader;
import java.net.HttpURLConnection;
import java.net.URL;
import java.text.NumberFormat;
import java.util.Timer;
import java.util.*;

public class Runner extends TimerTask {

    public final static String APP_NAME = "AJCollector";

    public final static ImageIcon appIcon = new ImageIcon(Runner.class.getResource("/resources/icon.png"));
    public final static ImageIcon appLogo = new ImageIcon(Runner.class.getResource("/resources/logo.png"));

    private final DocumentBuilderFactory dbf = DocumentBuilderFactory.newInstance();

    private final XPathFactory xpf = XPathFactory.newInstance();

    private final TreeMap<String, String> replacer = new TreeMap<>();

    private String coreVersion = "?";
    private String coreSystem = "?";

    private long coreCredits;
    private long coreConnections;
    private long coreSessionUpload;
    private long coreSessionDownload;
    private long coreUploadSpeed;
    private long coreDownloadSpeed;

    private long coreUploads;
    private long coreDownloads;
    private long coreDownloadsReady;

    private long shareFiles;
    private long sharedSize;

    private long networkUser;
    private long networkFiles;
    private double networkFileSize;

    public final Config config;

    private final TaskbarAndTray tt;

    private JFrame statusFrame;

    private DefaultTableModel statusModel;

    public static void main(String[] args) {
        new Runner();
    }

    public Runner() {
        config = new Config();

        tt = new TaskbarAndTray(this);

        Version version = new Version();

        version.check4update();

        Logger.info("Collector Version " + Version.getVersion());

        run();

        if (!GraphicsEnvironment.isHeadless()) {
            if (System.getProperty("os.name").toLowerCase().contains("linux")) {
                try {
                    Toolkit xToolkit = Toolkit.getDefaultToolkit();
                    java.lang.reflect.Field awtAppClassNameField = xToolkit.getClass().getDeclaredField("awtAppClassName");
                    awtAppClassNameField.setAccessible(true);
                    awtAppClassNameField.set(xToolkit, "AJCollector");
                } catch (Exception e) {
                    Logger.error(e);
                }
            }

            buildStatusFrame();
        }

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
            prepareShare();
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
            tt.setToolTip(InfoLine);
        }

        if (!GraphicsEnvironment.isHeadless()) {
            updateStatusFramePanel();
        }

        if (!config.getTargets().isEmpty()) {
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

        try {
            payload = Http.get(url);
        } catch (Exception e) {
            Logger.error(e.getMessage());
            return;
        }

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

    private void prepareShare() throws Exception {
        String url = String.format("%s:%s/xml/share.xml?password=%s", config.getCoreHost(), config.getCorePort(), config.getCorePassword());

        String payload;

        try {
            payload = Http.get(url);
        } catch (Exception e) {
            Logger.error(e.getMessage());
            return;
        }

        handleShare(payload);
    }

    private void handleShare(String payload) throws Exception {
        DocumentBuilder db = dbf.newDocumentBuilder();
        Document document = db.parse(new InputSource(new StringReader(payload)));

        XPath xpath = xpf.newXPath();
        Element shares = (Element) xpath.evaluate("/applejuice/shares", document, XPathConstants.NODE);

        shareFiles = 0;
        sharedSize = 0;
        for (int i = 0; i < shares.getElementsByTagName("share").getLength(); i++) {
            Element file = (Element) document.getElementsByTagName("share").item(i);
            shareFiles++;
            sharedSize += Long.parseLong(file.getAttribute("size"));
        }
    }

    private void prepareAppleJuiceInformation() throws Exception {
        String payload;

        String url = String.format("%s:%s/xml/modified.xml?password=%s", config.getCoreHost(), config.getCorePort(), config.getCorePassword());

        try {
            payload = Http.get(url);
        } catch (Exception e) {
            Logger.error(e.getMessage());
            return;
        }

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

        networkFileSize = Double.parseDouble(networkInfo.getAttribute("filesize").replace(",", "."));
    }

    private void forward() {
        config.getTargets().forEach(consumer -> {
            try {
                Target target = (Target) consumer;

                String forwardUrl = target.getUrl();
                String forwardToken = target.getToken();
                String forwardLine = target.getLine();

                String charset = "UTF-8";
                HttpURLConnection connection = (HttpURLConnection) new URL(forwardUrl).openConnection();
                connection.setDoOutput(true);
                connection.setRequestMethod("POST");
                connection.setRequestProperty("Accept-Charset", charset);
                connection.setRequestProperty("User-Agent", String.format("Collector/%s; Java/%s; (%s/%s)", Version.getVersion(), System.getProperty("java.version"), System.getProperty("os.name"), System.getProperty("os.version")));

                if (!forwardToken.isEmpty()) {
                    connection.setRequestProperty("Authorization", "Token " + forwardToken);
                }

                connection.setRequestProperty("Content-Type", "application/json;charset=" + charset);

                String Line = Json.object()
                        .add("forward_line", replacePlaceHolder(forwardLine))
                        .add("core_version", coreVersion)
                        .add("core_system", coreSystem)
                        .add("core_credits", coreCredits)
                        .add("core_connections", coreConnections)
                        .add("core_upload", coreUploadSpeed)
                        .add("core_download", coreDownloadSpeed)
                        .add("core_session_upload", coreSessionUpload)
                        .add("core_session_download", coreSessionDownload)
                        .add("core_uploads", coreUploads)
                        .add("core_downloads", coreDownloads)
                        .add("core_downloads_ready", coreDownloadsReady)
                        .add("share_files", shareFiles)
                        .add("share_size", sharedSize)
                        .add("network_user", networkUser)
                        .add("network_files", networkFiles)
                        .add("network_file_size", networkFileSize)
                        .toString();

                try (OutputStream output = connection.getOutputStream()) {
                    output.write(Line.getBytes(charset));
                }

                InputStream response = connection.getInputStream();
                Logger.info(String.format("forward successful: %s", forwardUrl));
            } catch (Exception e) {
                Logger.error(e.getMessage());
            }
        });
    }

    private String replacePlaceHolder(String Line) {
        for (Map.Entry<String, String> pair : replacer.entrySet()) {
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

        replacer.put("%shareFiles%", NumberFormat.getNumberInstance(Locale.GERMAN).format(shareFiles));
        replacer.put("%shareSize%", NumberFormatter.readableFileSize(sharedSize));

        replacer.put("%networkUser%", Long.toString(networkUser));
        replacer.put("%networkFiles%", NumberFormat.getNumberInstance(Locale.GERMAN).format(networkFiles));
        replacer.put("%networkFileSize%", NumberFormatter.readableNetworkShareSize(networkFileSize, 0));
    }

    private void buildStatusFrame() {
        statusFrame = new JFrame();
        statusFrame.setTitle(APP_NAME);
        statusFrame.setResizable(false);
        statusFrame.setSize(300, 280);
        statusFrame.setIconImage(appIcon.getImage());
        statusFrame.setAlwaysOnTop(true);

        JPanel pnlButton = new JPanel(new FlowLayout(FlowLayout.CENTER));

        statusFrame.addKeyListener(new KeyAdapter() {
            public void keyPressed(KeyEvent ke) {
                if (ke.getKeyCode() == KeyEvent.VK_ESCAPE) {
                    statusFrame.dispose();
                }
            }
        });

        statusModel = new DefaultTableModel(
                new Object[]{"Key", "Value"}, 0
        );

        JTable table = new JTable();
        table.setModel(statusModel);
        table.setEnabled(false);

        table.getTableHeader().setOpaque(false);

        JScrollPane scrollPane = new JScrollPane(table);
        scrollPane.setPreferredSize(new Dimension(299, 280));

        statusFrame.add(scrollPane, BorderLayout.CENTER);

        statusFrame.add(pnlButton, BorderLayout.PAGE_END);

        openStatusFrame();
    }

    private void updateStatusFramePanel() {
        if (statusFrame != null && statusFrame.isVisible()) {
            statusModel.setRowCount(0);
            for (Map.Entry<String, String> pair : replacer.entrySet()) {
                statusModel.addRow(new Object[]{pair.getKey(), pair.getValue()});
            }
        }
    }

    public void openStatusFrame() {
        statusFrame.pack();
        statusFrame.setLocationRelativeTo(null);
        statusFrame.setVisible(true);
    }
}
