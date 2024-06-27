package de.applejuicenet.collector;

import org.tinylog.Logger;
import org.w3c.dom.Document;
import org.w3c.dom.Element;
import org.w3c.dom.NodeList;

import javax.swing.*;
import javax.xml.parsers.DocumentBuilder;
import javax.xml.parsers.DocumentBuilderFactory;
import javax.xml.transform.OutputKeys;
import javax.xml.transform.Transformer;
import javax.xml.transform.TransformerFactory;
import javax.xml.transform.dom.DOMSource;
import javax.xml.transform.stream.StreamResult;
import java.awt.*;
import java.io.File;
import java.util.ArrayList;
import java.util.List;
import java.util.Objects;

class Config {

    private static final String FILENAME_XML = "core-information-collector.xml";

    private static final String DEFAULT_TRAYICON = "true";
    private static final String DEFAULT_TASKBARICON = "true";

    private static final String DEFAULT_INFO_LINE = "Credits %coreCredits% - Uploaded %coreSessionUpload% - Downloaded %coreSessionDownload% - Upload %coreUploadSpeed% - Download %coreDownloadSpeed%";
    private static final String DEFAULT_INTERVALL = "60000";

    private static final String DEFAULT_CORE_HOST = "http://127.0.0.1";
    private static final String DEFAULT_CORE_PORT = "9851";
    private static final String DEFAULT_CORE_PASSWD = "";

    private static final String DEFAULT_FORWARD_LINE = "Core `%coreVersion%` - Credits `%coreCredits%` - Uploaded `%coreSessionUpload%` - Downloaded `%coreSessionDownload%` - Upload `%coreUploadSpeed%` - Download `%coreDownloadSpeed%`";
    private static final String DEFAULT_FORWARD_URL = "https://discord.applejuicenet.cc/api/core-collector";
    private static final String DEFAULT_FORWARD_URL_OLD = "http://5f297e.online-server.cloud:82/api/core-collector";
    private static final String DEFAULT_FORWARD_TOKEN = "";

    private String trayIcon = DEFAULT_TRAYICON;
    private String taskbarIcon = DEFAULT_TASKBARICON;

    private String infoLine = DEFAULT_INFO_LINE;
    private String intervall = DEFAULT_INTERVALL;

    private String coreHost = DEFAULT_CORE_HOST;
    private String corePort = DEFAULT_CORE_PORT;
    private String corePassword = DEFAULT_CORE_PASSWD;

    private List<Target> targets = new ArrayList<Target>();

    public static File getConfigFile() {
        String rootDirectory = System.getProperty("user.home") + File.separator + "appleJuice" + File.separator + "collector";

        File aFile = new File(rootDirectory);

        if (!aFile.exists()) {
            aFile.mkdirs();
        }

        File fileXML = new File(rootDirectory + File.separator + Config.FILENAME_XML);

        if (!fileXML.exists()) {
            try {
                createConfig(
                        fileXML,
                        DEFAULT_TASKBARICON,
                        DEFAULT_INFO_LINE,
                        DEFAULT_INFO_LINE,
                        DEFAULT_INTERVALL,
                        DEFAULT_CORE_HOST,
                        DEFAULT_CORE_PORT,
                        DEFAULT_CORE_PASSWD,
                        DEFAULT_FORWARD_LINE,
                        DEFAULT_FORWARD_URL,
                        DEFAULT_FORWARD_TOKEN
                );
            } catch (Exception e) {
                Logger.error(e);
            }
        }

        return fileXML;
    }

    public Config() {
        File configFile = getConfigFile();

        try {
            readConfig(configFile);
        } catch (Exception e) {
            Logger.error(e);

            if (!GraphicsEnvironment.isHeadless()) {
                JOptionPane.showMessageDialog(null, e.getMessage(), Runner.APP_NAME, JOptionPane.ERROR_MESSAGE, Runner.appIcon);
            }

            System.exit(1);
        }
    }

    private void readConfig(File configFile) throws Exception {
        DocumentBuilderFactory documentBuilderFactory = DocumentBuilderFactory.newInstance();
        DocumentBuilder documentBuilder = documentBuilderFactory.newDocumentBuilder();
        Document document;

        document = documentBuilder.parse(configFile);
        document.getDocumentElement().normalize();

        infoLine = document.getElementsByTagName("infoLine").item(0).getTextContent();
        intervall = document.getDocumentElement().getAttribute("intervall");
        trayIcon = document.getDocumentElement().getAttribute("trayIcon");
        taskbarIcon = document.getDocumentElement().getAttribute("taskbarIcon");

        Element coreConfig = (Element) document.getElementsByTagName("core").item(0);

        coreHost = coreConfig.getAttribute("host");
        corePort = coreConfig.getAttribute("port");
        corePassword = coreConfig.getAttribute("password");

        NodeList forwardTargets = document.getElementsByTagName("target");

        for (int j = 0; j < forwardTargets.getLength(); j++) {
            Element elem = (Element) forwardTargets.item(j);
            String forwardUrl = elem.getElementsByTagName("url").item(0).getTextContent();

            if (forwardUrl.equals(DEFAULT_FORWARD_URL_OLD)) {
                forwardUrl = DEFAULT_FORWARD_URL;
            }

            String forwardToken = elem.getElementsByTagName("token").item(0).getTextContent();
            String forwardLine = elem.getElementsByTagName("line").item(0).getTextContent();

            targets.add(new Target(forwardUrl, forwardToken, forwardLine));
        }
    }

    public Boolean isTrayIcon() {
        return Objects.equals(trayIcon, "true") || Objects.equals(trayIcon, "");
    }

    public Boolean isTaskBarIcon() {
        return Objects.equals(taskbarIcon, "true") || Objects.equals(taskbarIcon, "");
    }

    public String getInfoLine() {
        return infoLine;
    }

    public Integer getInterval() {
        return Integer.parseInt(intervall);
    }

    public String getCoreHost() {
        return coreHost;
    }

    public String getCorePort() {
        return corePort;
    }

    public String getCorePassword() {
        return !corePassword.equals("") ? corePassword : "d41d8cd98f00b204e9800998ecf8427e";
    }

    public List getTargets() {
        return targets;
    }

    private static void createConfig(File fileXML, String trayIcon, String taskbarIcon, String infoLine, String intervall, String host, String port, String password, String forwardLine, String forwardUrl, String forwardToken) throws Exception {
        Document doc = DocumentBuilderFactory.newInstance().newDocumentBuilder().newDocument();
        Element root = doc.createElement("collector");
        root.setAttribute("intervall", intervall);
        root.setAttribute("trayIcon", trayIcon);
        root.setAttribute("taskbarIcon", taskbarIcon);

        Element info = doc.createElement("infoLine");
        info.appendChild(doc.createTextNode(infoLine));
        root.appendChild(info);

        Element core = doc.createElement("core");

        core.setAttribute("host", host);
        core.setAttribute("port", port);
        core.setAttribute("password", password);

        root.appendChild(core);

        doc.appendChild(root);

        Element targets = doc.createElement("targets");

        Element target = doc.createElement("target");

        Element targetURL = doc.createElement("url");
        targetURL.appendChild(doc.createTextNode(forwardUrl));
        target.appendChild(targetURL);

        Element targetToken = doc.createElement("token");
        targetToken.appendChild(doc.createTextNode(forwardToken));
        target.appendChild(targetToken);

        Element targetLine = doc.createElement("line");
        targetLine.appendChild(doc.createTextNode(forwardLine));
        target.appendChild(targetLine);

        targets.appendChild(target);

        root.appendChild(targets);

        DOMSource domSource = new DOMSource(doc);

        Transformer transformer = TransformerFactory.newInstance().newTransformer();
        transformer.setOutputProperty(OutputKeys.INDENT, "yes");
        transformer.setOutputProperty("{http://xml.apache.org/xslt}indent-amount", "2");

        StreamResult streamResult = new StreamResult(fileXML);
        transformer.transform(domSource, streamResult);
    }
}

