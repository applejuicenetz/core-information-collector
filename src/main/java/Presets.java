import java.io.*;
import java.util.Optional;
import java.util.Properties;

class Presets {
    private final String FILENAME = "core-information-collector.properties";

    private Properties config;

    private static final String DEFAULT_INFO_LINE = "Credits %coreCredits% - Uploaded %coreSessionUpload% - Downloaded %coreSessionDownload% - Upload %coreUploadSpeed% - Download %coreDownloadSpeed%";
    private static final String DEFAULT_FORWARD_LINE = "Core `%coreVersion%` - Credits `%coreCredits%` - Uploaded `%coreSessionUpload%` - Downloaded `%coreSessionDownload%` - Upload `%coreUploadSpeed%` - Download `%coreDownloadSpeed%`";
    private static final String DEFAULT_FORWARD_URL = "https://discord-bot.knastbruder.applejuicenet.de/api/core-collector";
    private static final String DEFAULT_FORWARD_TOKEN = "";
    private static final String DEFAULT_INTERVAL = "60000";
    private static final String DEFAULT_CORE_HOST = "http://127.0.0.1";
    private static final String DEFAULT_CORE_PORT = "9851";
    private static final String DEFAULT_CORE_PASSWD = "";

    Presets() {
        config = new Properties();

        // for docker container
        if (!Optional.ofNullable(System.getenv("CORE_HOST")).orElse("").isEmpty()) {
            config.setProperty("info_line", Optional.ofNullable(System.getenv("INFO_LINE")).orElse(DEFAULT_INFO_LINE));
            config.setProperty("forward_line", Optional.ofNullable(System.getenv("FORWARD_LINE")).orElse(DEFAULT_FORWARD_LINE));
            config.setProperty("forward_url", Optional.ofNullable(System.getenv("FORWARD_URL")).orElse(DEFAULT_FORWARD_URL));
            config.setProperty("forward_token", Optional.ofNullable(System.getenv("FORWARD_TOKEN")).orElse(DEFAULT_FORWARD_TOKEN));
            config.setProperty("interval", Optional.ofNullable(System.getenv("INTERVAL")).orElse(DEFAULT_INTERVAL));
            config.setProperty("core_host", Optional.ofNullable(System.getenv("CORE_HOST")).orElse(DEFAULT_CORE_HOST));
            config.setProperty("core_port", Optional.ofNullable(System.getenv("CORE_PORT")).orElse(DEFAULT_CORE_PORT));
            config.setProperty("core_passwd", Optional.ofNullable(System.getenv("CORE_PASSWD")).orElse(DEFAULT_CORE_PASSWD));
        } else {
            File file = new File(this.FILENAME);

            if (file.exists()) {
                loadPresets();
            } else {
                config.setProperty("info_line", DEFAULT_INFO_LINE);
                config.setProperty("forward_line", DEFAULT_FORWARD_LINE);
                config.setProperty("forward_url", DEFAULT_FORWARD_URL);
                config.setProperty("forward_token", DEFAULT_FORWARD_TOKEN);
                config.setProperty("interval", DEFAULT_INTERVAL);
                config.setProperty("core_host", DEFAULT_CORE_HOST);
                config.setProperty("core_port", DEFAULT_CORE_PORT);
                config.setProperty("core_passwd", DEFAULT_CORE_PASSWD);

                savePresets();
            }
        }
    }

    public String getInfoLine() {
        return config.getProperty("info_line", DEFAULT_INFO_LINE);
    }

    public String getForwardLine() {
        return config.getProperty("forward_line", DEFAULT_FORWARD_LINE);
    }

    public String getForwardUrl() {
        return config.getProperty("forward_url", DEFAULT_FORWARD_URL);
    }

    public String getForwardToken() {
        return config.getProperty("forward_token", DEFAULT_FORWARD_TOKEN);
    }

    public Integer getInterval() {
        return Integer.parseInt(config.getProperty("interval", DEFAULT_INTERVAL));
    }

    public String getCoreHost() {
        return config.getProperty("core_host", DEFAULT_CORE_HOST);
    }

    public String getCorePort() {
        return config.getProperty("core_port", DEFAULT_CORE_PORT);
    }

    public String getCorePassword() {
        return config.getProperty("core_passwd", DEFAULT_CORE_PASSWD);
    }

    private void loadPresets() {
        try {
            InputStream fis = new FileInputStream(this.FILENAME);
            config.load(fis);
        } catch (IOException ex) {
            ex.printStackTrace();
        }
    }

    private void savePresets() {
        try (OutputStream outputStream = new FileOutputStream(this.FILENAME)) {
            config.store(outputStream, null);
        } catch (IOException e) {
            System.err.println(e.toString());
        }
    }
}
