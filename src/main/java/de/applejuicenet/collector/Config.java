package de.applejuicenet.collector;

import org.tinylog.Logger;
import org.w3c.dom.Document;
import org.w3c.dom.Element;
import org.w3c.dom.NodeList;
import org.xml.sax.SAXException;

import javax.swing.*;
import javax.xml.parsers.DocumentBuilder;
import javax.xml.parsers.DocumentBuilderFactory;
import javax.xml.parsers.ParserConfigurationException;
import javax.xml.transform.OutputKeys;
import javax.xml.transform.Transformer;
import javax.xml.transform.TransformerFactory;
import javax.xml.transform.dom.DOMSource;
import javax.xml.transform.stream.StreamResult;
import java.io.File;
import java.io.FileInputStream;
import java.io.IOException;
import java.io.InputStream;
import java.nio.file.Files;
import java.nio.file.Paths;
import java.nio.file.StandardCopyOption;
import java.util.ArrayList;
import java.util.List;
import java.util.Properties;

class Config {
    private static String rootDirectory = null;

    private final String FILENAME_PROPS = "core-information-collector.properties";
    private final String FILENAME_XML = "core-information-collector.xml";

    private static final String DEFAULT_INFO_LINE = "Credits %coreCredits% - Uploaded %coreSessionUpload% - Downloaded %coreSessionDownload% - Upload %coreUploadSpeed% - Download %coreDownloadSpeed%";
    private static final String DEFAULT_INTERVALL = "60000";

    private static final String DEFAULT_CORE_HOST = "http://127.0.0.1";
    private static final String DEFAULT_CORE_PORT = "9851";
    private static final String DEFAULT_CORE_PASSWD = "";

    private static final String DEFAULT_FORWARD_LINE = "Core `%coreVersion%` - Credits `%coreCredits%` - Uploaded `%coreSessionUpload%` - Downloaded `%coreSessionDownload%` - Upload `%coreUploadSpeed%` - Download `%coreDownloadSpeed%`";
    private static final String DEFAULT_FORWARD_URL = "http://5f297e.online-server.cloud:82/api/core-collector";
    private static final String DEFAULT_FORWARD_TOKEN = "";

    private String infoLine = DEFAULT_INFO_LINE;
    private String intervall = DEFAULT_INTERVALL;

    private String coreHost = DEFAULT_CORE_HOST;
    private String corePort = DEFAULT_CORE_PORT;
    private String corePassword = DEFAULT_CORE_PASSWD;

    private List<Target> targets = new ArrayList<Target>();

    public Config() {
        rootDirectory = System.getProperty("user.home") + File.separator + "appleJuice" + File.separator + "collector";

        File aFile = new File(rootDirectory);

        if (!aFile.exists()) {
            aFile.mkdir();
        }

        try {
            moveConfig();
        } catch (IOException e) {
            Logger.error(e);
        }

        File fileProps = new File(rootDirectory + File.separator + this.FILENAME_PROPS);
        File fileXML = new File(rootDirectory + File.separator + this.FILENAME_XML);

        if (fileProps.exists() && !fileXML.exists()) {
            convertPropertiestoXML();
        }

        if (!fileXML.exists()) {
            try {
                createConfig(
                        DEFAULT_INFO_LINE,
                        DEFAULT_INTERVALL,
                        DEFAULT_CORE_HOST,
                        DEFAULT_CORE_PORT,
                        DEFAULT_CORE_PASSWD,
                        DEFAULT_FORWARD_LINE,
                        DEFAULT_FORWARD_URL,
                        DEFAULT_FORWARD_TOKEN
                );
                fileProps.delete();
            } catch (Exception e) {
                Logger.error(e);
            }
        }

        try {
            readConfig();
        } catch (Exception e) {
            Logger.error(e);

            JOptionPane.showMessageDialog(null, e.getMessage(), Runner.APP_NAME, JOptionPane.ERROR_MESSAGE, Runner.appIcon);

            System.exit(1);
        }
    }

    private void moveConfig() throws IOException {
        File fileProps = new File(this.FILENAME_PROPS);
        File fileXML = new File(this.FILENAME_XML);

        if (fileProps.exists()) {
            Files.move(Paths.get(this.FILENAME_PROPS), Paths.get(rootDirectory + File.separator + this.FILENAME_PROPS), StandardCopyOption.REPLACE_EXISTING);
        }

        if (fileXML.exists()) {
            Files.move(Paths.get(this.FILENAME_XML), Paths.get(rootDirectory + File.separator + this.FILENAME_XML), StandardCopyOption.REPLACE_EXISTING);
        }
    }

    private void readConfig() throws Exception {
        File file = new File(rootDirectory + File.separator + FILENAME_XML);
        DocumentBuilderFactory documentBuilderFactory = DocumentBuilderFactory.newInstance();
        DocumentBuilder documentBuilder = documentBuilderFactory.newDocumentBuilder();
        Document document = null;

        document = documentBuilder.parse(file);
        document.getDocumentElement().normalize();

        infoLine = document.getElementsByTagName("infoLine").item(0).getTextContent();
        intervall = document.getDocumentElement().getAttribute("intervall");

        Element coreConfig = (Element) document.getElementsByTagName("core").item(0);

        coreHost = coreConfig.getAttribute("host");
        corePort = coreConfig.getAttribute("port");
        corePassword = coreConfig.getAttribute("password");

        NodeList forwardTargets = document.getElementsByTagName("target");

        for (int j = 0; j < forwardTargets.getLength(); j++) {
            Element elem = (Element) forwardTargets.item(j);
            String forwardUrl = elem.getElementsByTagName("url").item(0).getTextContent();
            String forwardToken = elem.getElementsByTagName("token").item(0).getTextContent();
            String forwardLine = elem.getElementsByTagName("line").item(0).getTextContent();

            targets.add(new Target(forwardUrl, forwardToken, forwardLine));
        }
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
        return corePassword;
    }

    public List getTargets() {
        return targets;
    }

    private void convertPropertiestoXML() {
        Properties props = new Properties();

        try {
            InputStream fis = new FileInputStream(this.FILENAME_PROPS);
            props.load(fis);
        } catch (IOException ex) {
            ex.printStackTrace();
        }
        try {
            createConfig(
                    props.getProperty("info_line"),
                    props.getProperty("interval"),
                    props.getProperty("core_host"),
                    props.getProperty("core_port"),
                    props.getProperty("core_passwd"),
                    props.getProperty("forward_line"),
                    props.getProperty("forward_url"),
                    props.getProperty("forward_token")
            );
        } catch (Exception e) {
            Logger.error(e);
        }
    }

    private void createConfig(String infoLine, String intervall, String host, String port, String password, String forwardLine, String forwardUrl, String forwardToken) throws Exception {
        Document doc = DocumentBuilderFactory.newInstance().newDocumentBuilder().newDocument();
        Element root = doc.createElement("collector");
        root.setAttribute("intervall", intervall);

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

        StreamResult streamResult = new StreamResult(new File(FILENAME_XML));
        transformer.transform(domSource, streamResult);
    }
}
