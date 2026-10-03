package de.applejuicenet.collector;

import java.io.BufferedInputStream;
import java.io.BufferedReader;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.net.HttpURLConnection;
import java.net.URI;
import java.nio.charset.StandardCharsets;

public class Http {
    private static final String WRONG_PASSWORD = "wrong password. access denied.";

    public static final int DEFAULT_TIMEOUT = 5000;

    @FunctionalInterface
    public interface StreamHandler<T> {
        T handle(InputStream stream) throws Exception;
    }

    public static String get(String urlToRead) throws Exception {
        return get(urlToRead, DEFAULT_TIMEOUT);
    }

    public static String get(String urlToRead, int timeout) throws Exception {
        String payload = stream(urlToRead, timeout, stream -> {
            StringBuilder result = new StringBuilder();
            try (BufferedReader rd = new BufferedReader(new InputStreamReader(stream, StandardCharsets.UTF_8))) {
                String line;
                while ((line = rd.readLine()) != null) {
                    result.append(line);
                }
            }
            return result.toString();
        });

        if (payload.contains(WRONG_PASSWORD)) {
            throw new Exception(WRONG_PASSWORD);
        }

        return payload;
    }

    public static <T> T stream(String urlToRead, int timeout, StreamHandler<T> handler) throws Exception {
        HttpURLConnection conn = (HttpURLConnection) URI.create(urlToRead).toURL().openConnection();
        try {
            conn.setRequestMethod("GET");
            conn.setConnectTimeout(timeout);
            conn.setReadTimeout(timeout);

            try (InputStream in = new BufferedInputStream(conn.getInputStream())) {
                return handler.handle(in);
            }
        } catch (Exception e) {
            String message = e.getMessage() != null ? e.getMessage() : e.toString();
            throw new Exception(maskPassword(message));
        } finally {
            conn.disconnect();
        }
    }

    static String maskPassword(String message) {
        return message.replaceAll("password=[^&\\s]*", "password=***");
    }

    public static void failOnWrongPassword(BufferedInputStream in) throws Exception {
        in.mark(128);
        byte[] head = in.readNBytes(128);
        in.reset();

        if (new String(head, StandardCharsets.UTF_8).contains(WRONG_PASSWORD)) {
            throw new Exception(WRONG_PASSWORD);
        }
    }
}
