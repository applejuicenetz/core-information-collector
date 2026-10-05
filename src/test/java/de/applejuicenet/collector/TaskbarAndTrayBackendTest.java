package de.applejuicenet.collector;

import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertTrue;

class TaskbarAndTrayBackendTest {
    @Test
    void skipsNativeBackendOn32BitJvm() {
        assertFalse(TaskbarAndTray.useNativeTray("Linux", "32"));
    }

    @Test
    void keepsNativeBackendOn64BitLinuxOnly() {
        assertTrue(TaskbarAndTray.useNativeTray("Linux", "64"));
        assertFalse(TaskbarAndTray.useNativeTray("Windows 11", "64"));
        assertFalse(TaskbarAndTray.useNativeTray("Mac OS X", "64"));
        assertFalse(TaskbarAndTray.useNativeTray("Linux", ""));
    }
}
