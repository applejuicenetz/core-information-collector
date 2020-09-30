import java.io.*;
import java.util.Properties;

class Presets {
    private final String FILENAME = "core-information-collector.properties";

    private Properties config;

    private static final String DEFAULT_INFOLINE = "Credits %coreCredits% - Uploaded %coreSessionUpload% - Downloaded %coreSessionDownload% - Upload %coreUploadSpeed% - Download %coreDownloadSpeed%";
    private static final String DEFAULT_COLLECTOR = "https://discord-bot.knastbruder.applejuicent.de/api/core-collector";
    private static final String DEFAULT_COLLECTOR_TOKEN = "";
    private static final String DEFAULT_INTERVAL = "60000";
    private static final String DEFAULT_CORE_HOST = "http://127.0.0.1";
    private static final String DEFAULT_CORE_PORT = "9851";
    private static final String DEFAULT_CORE_PASSWD = "";

    Presets() {
        File file = new File(this.FILENAME);
        config = new Properties();

        if (file.exists()) {
            loadPresets();
        } else {
            config.setProperty("info_line", DEFAULT_INFOLINE);
            config.setProperty("collector_url", DEFAULT_COLLECTOR);
            config.setProperty("collector_token", DEFAULT_COLLECTOR_TOKEN);
            config.setProperty("interval", DEFAULT_INTERVAL);
            config.setProperty("core_host", DEFAULT_CORE_HOST);
            config.setProperty("core_port", DEFAULT_CORE_PORT);
            config.setProperty("core_passwd", DEFAULT_CORE_PASSWD);

            savePresets();
        }
    }

    public String getInfoLine() {
        return config.getProperty("info_line", DEFAULT_INFOLINE);
    }

    public String getCollector() {
        return config.getProperty("collector", DEFAULT_COLLECTOR_TOKEN);
    }

    public String getCollectorToken() {
        return config.getProperty("collector_token", DEFAULT_COLLECTOR);
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
