package de.applejuicenet.collector;

import com.eclipsesource.json.Json;
import com.eclipsesource.json.JsonObject;
import org.tinylog.Logger;

import javax.swing.*;
import java.awt.*;
import java.net.URI;
import java.util.Arrays;
import java.util.List;
import java.util.stream.Collectors;

public class Version {
    private static final String GITHUB_API_URL = "https://api.github.com/repos/applejuicenetz/core-information-collector/releases/latest";
    private static final String GITHUB_URL = "https://github.com/applejuicenetz/core-information-collector/releases";

    private final Runner runner;

    public Version(Runner runner) {
        this.runner = runner;
    }

    public void check4update() {
        String payload;
        try {
            payload = Http.get(GITHUB_API_URL);
        } catch (Exception e) {
            Logger.error(e);
            return;
        }

        JsonObject obj = Json.parse(payload).asObject();

        final String currentVersion = getVersion();
        final String newVersion = obj.getString("tag_name", "0.0.0");

        if (!currentVersion.contains("SNAPSHOT") && compareVersion(newVersion, currentVersion) > 0) {
            final String updateMessage = "neue Version %newVersion% verfügbar, download Seite öffnen?".replace("%newVersion%", newVersion);
            int input = JOptionPane.showConfirmDialog(
                    null,
                    updateMessage,
                    runner.APP_NAME,
                    JOptionPane.YES_NO_OPTION,
                    JOptionPane.INFORMATION_MESSAGE,
                    runner.appIcon);

            if (0 == input) {
                openUpdatenUrl();
            }
        }
    }

    public static String getVersion() {
        String currentVersion = Version.class.getPackage().getImplementationVersion();
        return currentVersion != null ? currentVersion : "SNAPSHOT";
    }

    private void openUpdatenUrl() {
        try {
            Desktop.getDesktop().browse(new URI(GITHUB_URL));
        } catch (Exception e) {
            Logger.error(e);
        }
    }

    public int compareVersion(String A, String B) {
        java.util.List<String> strList1 = Arrays.stream(A.split("\\."))
                .map(s -> s.replaceAll("^0+(?!$)", ""))
                .collect(Collectors.toList());
        List<String> strList2 = Arrays.stream(B.split("\\."))
                .map(s -> s.replaceAll("^0+(?!$)", ""))
                .collect(Collectors.toList());
        int len1 = strList1.size();
        int len2 = strList2.size();
        int i = 0;
        while (i < len1 && i < len2) {
            if (strList1.get(i).length() > strList2.get(i).length()) return 1;
            if (strList1.get(i).length() < strList2.get(i).length()) return -1;
            int result = new Long(strList1.get(i)).compareTo(new Long(strList2.get(i)));
            if (result != 0) return result;
            i++;
        }
        while (i < len1) {
            if (!strList1.get(i++).equals("0")) return 1;
        }
        while (i < len2) {
            if (!strList2.get(i++).equals("0")) return -1;
        }
        return 0;
    }
}
