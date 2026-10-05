package de.applejuicenet.collector;

import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.io.TempDir;

import java.nio.file.Files;
import java.nio.file.Path;
import static org.junit.jupiter.api.Assertions.*;

class ConfigTest {
    @TempDir Path home;
    private String originalHome;

    @BeforeEach void isolateHome() {
        originalHome = System.getProperty("user.home");
        System.setProperty("user.home", home.toString());
    }

    @AfterEach void restoreHome() {
        System.setProperty("user.home", originalHome);
    }

    private Config config(String interval, String timeout, String password) throws Exception {
        Path folder = home.resolve("appleJuice/collector");
        Files.createDirectories(folder);
        Files.writeString(folder.resolve("collector.xml"),
                "<collector intervall=\"" + interval + "\" trayIcon=\"false\" taskbarIcon=\"false\">"
                + "<infoLine>test</infoLine><core host=\"http://127.0.0.1\" port=\"9851\" password=\""
                + password + "\" timeout=\"" + timeout + "\"/></collector>");
        return new Config();
    }

    @Test void acceptsExplicitIntervalAndTimeout() throws Exception {
        Config config = config("12345", "2345", "test-hash");
        assertEquals(12345, config.getInterval());
        assertEquals(2345, config.getCoreTimeout());
        assertEquals("test-hash", config.getCorePassword());
        assertFalse(config.isTrayIcon());
        assertTrue(config.getTargets().isEmpty());
    }

    @Test void invalidNumbersFallBackToDefaults() throws Exception {
        for (String value : new String[]{"", "0", "-1", "bad", "2147483648"}) {
            Config config = config(value, value, "");
            assertEquals(60000, config.getInterval());
            assertEquals(Http.DEFAULT_TIMEOUT, config.getCoreTimeout());
        }
    }

    @Test void emptyPasswordUsesEmptyMd5Hash() throws Exception {
        assertEquals("d41d8cd98f00b204e9800998ecf8427e", config("5000", "5000", "").getCorePassword());
    }

    @Test void generatedDefaultsEnableTrayAndTaskbar() {
        Config config = new Config();
        assertTrue(config.isTrayIcon());
        assertTrue(config.isTaskBarIcon());
        assertEquals(60000, config.getInterval());
        assertTrue(Config.getConfigFile().isFile());
    }
}
