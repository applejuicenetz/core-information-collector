package de.applejuicenet.collector;

import java.io.*;
import java.util.Optional;
import java.util.Properties;

class Config {
    private final String FILENAME = "core-information-collector.properties";

    private final Properties props;

    private static final String DEFAULT_INFO_LINE = "Credits %coreCredits% - Uploaded %coreSessionUpload% - Downloaded %coreSessionDownload% - Upload %coreUploadSpeed% - Download %coreDownloadSpeed%";
    private static final String DEFAULT_FORWARD_LINE = "Core `%coreVersion%` - Credits `%coreCredits%` - Uploaded `%coreSessionUpload%` - Downloaded `%coreSessionDownload%` - Upload `%coreUploadSpeed%` - Download `%coreDownloadSpeed%`";
    private static final String DEFAULT_FORWARD_URL = "https://discord.applejuicenet.cc/api/core-collector";
    private static final String DEFAULT_FORWARD_TOKEN = "";
    private static final String DEFAULT_INTERVAL = "60000";
    private static final String DEFAULT_CORE_HOST = "http://127.0.0.1";
    private static final String DEFAULT_CORE_PORT = "9851";
    private static final String DEFAULT_CORE_PASSWD = "";

    public Config() {
        props = new Properties();

        // for docker container
        if (!Optional.ofNullable(System.getenv("CORE_HOST")).orElse("").isEmpty()) {
            props.setProperty("info_line", Optional.ofNullable(System.getenv("INFO_LINE")).orElse(DEFAULT_INFO_LINE));
            props.setProperty("forward_line", Optional.ofNullable(System.getenv("FORWARD_LINE")).orElse(DEFAULT_FORWARD_LINE));
            props.setProperty("forward_url", Optional.ofNullable(System.getenv("FORWARD_URL")).orElse(DEFAULT_FORWARD_URL));
            props.setProperty("forward_token", Optional.ofNullable(System.getenv("FORWARD_TOKEN")).orElse(DEFAULT_FORWARD_TOKEN));
            props.setProperty("interval", Optional.ofNullable(System.getenv("INTERVAL")).orElse(DEFAULT_INTERVAL));
            props.setProperty("core_host", Optional.ofNullable(System.getenv("CORE_HOST")).orElse(DEFAULT_CORE_HOST));
            props.setProperty("core_port", Optional.ofNullable(System.getenv("CORE_PORT")).orElse(DEFAULT_CORE_PORT));
            props.setProperty("core_passwd", Optional.ofNullable(System.getenv("CORE_PASSWD")).orElse(DEFAULT_CORE_PASSWD));
        } else {
            File file = new File(this.FILENAME);

            if (file.exists()) {
                loadPresets();
            } else {
                props.setProperty("info_line", DEFAULT_INFO_LINE);
                props.setProperty("forward_line", DEFAULT_FORWARD_LINE);
                props.setProperty("forward_url", DEFAULT_FORWARD_URL);
                props.setProperty("forward_token", DEFAULT_FORWARD_TOKEN);
                props.setProperty("interval", DEFAULT_INTERVAL);
                props.setProperty("core_host", DEFAULT_CORE_HOST);
                props.setProperty("core_port", DEFAULT_CORE_PORT);
                props.setProperty("core_passwd", DEFAULT_CORE_PASSWD);

                savePresets();
            }
        }
    }

    public String getInfoLine() {
        return props.getProperty("info_line", DEFAULT_INFO_LINE);
    }

    public String getForwardLine() {
        return props.getProperty("forward_line", DEFAULT_FORWARD_LINE);
    }

    public String getForwardUrl() {
        return props.getProperty("forward_url", DEFAULT_FORWARD_URL);
    }

    public String getForwardToken() {
        return props.getProperty("forward_token", DEFAULT_FORWARD_TOKEN);
    }

    public Integer getInterval() {
        return Integer.parseInt(props.getProperty("interval", DEFAULT_INTERVAL));
    }

    public String getCoreHost() {
        return props.getProperty("core_host", DEFAULT_CORE_HOST);
    }

    public String getCorePort() {
        return props.getProperty("core_port", DEFAULT_CORE_PORT);
    }

    public String getCorePassword() {
        return props.getProperty("core_passwd", DEFAULT_CORE_PASSWD);
    }

    private void loadPresets() {
        try {
            InputStream fis = new FileInputStream(this.FILENAME);
            props.load(fis);
        } catch (IOException ex) {
            ex.printStackTrace();
        }
    }

    private void savePresets() {
        try (OutputStream outputStream = new FileOutputStream(this.FILENAME)) {
            props.store(outputStream, null);
        } catch (IOException e) {
            System.err.println(e.toString());
        }
    }
}
