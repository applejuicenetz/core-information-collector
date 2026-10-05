package de.applejuicenet.collector;

import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.assertEquals;

class NumberFormatterTest {
    @Test
    void zeroAndNegativeSizesAreZero() {
        assertEquals("0", NumberFormatter.readableFileSize(0));
        assertEquals("0", NumberFormatter.readableFileSize(-1));
    }

    @Test
    void usesBinaryUnitsAtBoundaries() {
        assertEquals("1B", NumberFormatter.readableFileSize(1));
        assertEquals(new java.text.DecimalFormat("#,##0.##").format(1023) + "B",
                NumberFormatter.readableFileSize(1023));
        assertEquals("1kB", NumberFormatter.readableFileSize(1024));
        assertEquals("1MB", NumberFormatter.readableFileSize(1024L * 1024));
        assertEquals("1GB", NumberFormatter.readableFileSize(1024L * 1024 * 1024));
    }

    @Test
    void handlesFilesLargerThanTwoGibWithoutOverflow() {
        assertEquals("2GB", NumberFormatter.readableFileSize(2L * 1024 * 1024 * 1024));
        assertEquals("4GB", NumberFormatter.readableFileSize(4L * 1024 * 1024 * 1024));
        assertEquals("1TB", NumberFormatter.readableFileSize(1024L * 1024 * 1024 * 1024));
        assertEquals("8PB", NumberFormatter.readableFileSize(8L * 1024 * 1024 * 1024 * 1024 * 1024));
    }

    @Test
    void largestLongSizeDoesNotExceedAvailableUnits() {
        assertEquals(new java.text.DecimalFormat("#,##0.##").format(
                Long.MAX_VALUE / Math.pow(1024, 5)) + "PB",
                NumberFormatter.readableFileSize(Long.MAX_VALUE));
    }

    @Test
    void networkSizeChoosesUnitAndUsesComma() {
        assertEquals("0,00 MB", NumberFormatter.readableNetworkShareSize(0, 0));
        assertEquals("1,5GB", NumberFormatter.readableNetworkShareSize(1536, 0));
        assertEquals("2,25TB", NumberFormatter.readableNetworkShareSize(2.25 * 1024 * 1024, 0));
    }

    @Test
    void explicitNetworkFactorsMapToUnits() {
        assertEquals("2,0MB", NumberFormatter.readableNetworkShareSize(2, 1));
        assertEquals("2,0GB", NumberFormatter.readableNetworkShareSize(2048, 1024));
        assertEquals("0,14??", NumberFormatter.readableNetworkShareSize(1, 7));
    }
}
