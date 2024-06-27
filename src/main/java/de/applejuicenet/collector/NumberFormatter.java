package de.applejuicenet.collector;

import java.text.DecimalFormat;

public class NumberFormatter {

    public static String readableFileSize(long size) {
        if (size <= 0) return "0";
        final String[] units = new String[]{"B", "kB", "MB", "GB", "TB", "PB"};
        int digitGroups = (int) (Math.log10(size) / Math.log10(1024));
        return new DecimalFormat("#,##0.##").format(size / Math.pow(1024, digitGroups)) + units[digitGroups];
    }

    public static String readableNetworkShareSize(double share, long faktor) {

        if (share == 0) {
            return "0,00 MB";
        }

        if (faktor == 0) { // selbst entscheiden
            if (share / 1024 < 1024) {
                faktor = 1024;
            } else if (share / 1048576 < 1024) {
                faktor = 1048576;
            } else if (share / 1073741824 < 1024) {
                faktor = 1073741824;
            } else {
                faktor = 1;
            }
        }

        share = share / faktor;
        String result = Double.toString(share);

        if (result.indexOf(".") + 3 < result.length()) {
            result = result.substring(0, result.indexOf(".") + 3);
        }

        result = result.replace('.', ',');
        if (faktor == 1) {
            result += "MB";
        } else if (faktor == 1024) {
            result += "GB";
        } else if (faktor == 1048576) {
            result += "TB";
        } else if (faktor == 1073741824) {
            result += "PB";
        } else {
            result += "??";
        }

        return result;
    }
}
