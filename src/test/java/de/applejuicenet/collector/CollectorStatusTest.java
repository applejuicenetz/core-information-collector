package de.applejuicenet.collector;

import org.junit.jupiter.api.Test;
import java.util.Map;
import static org.junit.jupiter.api.Assertions.*;

class CollectorStatusTest {
    @Test
    void failureShowsStatusWithoutStaleCoreValues() {
        Map<String, String> rows = Runner.statusSnapshot(Map.of("%coreCredits%", "123"), false);
        assertEquals("Core-Abfrage fehlgeschlagen", rows.get("Status"));
        assertFalse(rows.containsKey("%coreCredits%"));
    }

    @Test
    void successfulSnapshotKeepsCollectedValues() {
        Map<String, String> rows = Runner.statusSnapshot(Map.of("%coreCredits%", "123"), true);
        assertEquals("Core verbunden", rows.get("Status"));
        assertEquals("123", rows.get("%coreCredits%"));
    }
}
