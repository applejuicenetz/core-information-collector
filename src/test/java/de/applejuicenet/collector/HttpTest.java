package de.applejuicenet.collector;

import com.sun.net.httpserver.HttpServer;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import java.io.BufferedInputStream;
import java.io.ByteArrayInputStream;
import java.io.InputStream;
import java.net.InetSocketAddress;
import java.nio.charset.StandardCharsets;
import java.util.concurrent.atomic.AtomicReference;

import static org.junit.jupiter.api.Assertions.*;

class HttpTest {
    private HttpServer server;
    private String url;

    @BeforeEach
    void startServer() throws Exception {
        server = HttpServer.create(new InetSocketAddress("127.0.0.1", 0), 0);
        url = "http://127.0.0.1:" + server.getAddress().getPort();
    }

    @AfterEach
    void stopServer() {
        server.stop(0);
    }

    private void response(String path, int status, String text) {
        server.createContext(path, exchange -> {
            byte[] bytes = text.getBytes(StandardCharsets.UTF_8);
            exchange.sendResponseHeaders(status, bytes.length);
            try (var output = exchange.getResponseBody()) {
                output.write(bytes);
            }
        });
        server.start();
    }

    @Test
    void readsUtf8Response() throws Exception {
        response("/data", 200, "Größe: 3 GiB\nÄpfel");
        assertEquals("Größe: 3 GiBÄpfel", Http.get(url + "/data", 2000));
    }

    @Test
    void rejectsWrongPasswordResponse() {
        response("/denied", 200, "wrong password. access denied.");
        Exception failure = assertThrows(Exception.class, () -> Http.get(url + "/denied", 2000));
        assertEquals("wrong password. access denied.", failure.getMessage());
    }

    @Test
    void masksPasswordInHttpError() {
        response("/failure", 500, "failure");
        Exception failure = assertThrows(Exception.class,
                () -> Http.get(url + "/failure?password=private-value&other=1", 2000));
        assertFalse(failure.getMessage().contains("private-value"));
        assertTrue(failure.getMessage().contains("password=***"));
    }

    @Test
    void masksEveryPasswordAndPreservesOtherParameters() {
        assertEquals("password=***&x=1 password=*** end",
                Http.maskPassword("password=first&x=1 password=second end"));
    }

    @Test
    void wrongPasswordProbeLeavesSuccessfulStreamUnchanged() throws Exception {
        byte[] payload = ("<applejuice>" + "x".repeat(256) + "</applejuice>")
                .getBytes(StandardCharsets.UTF_8);
        try (var input = new BufferedInputStream(new ByteArrayInputStream(payload))) {
            Http.failOnWrongPassword(input);
            assertArrayEquals(payload, input.readAllBytes());
        }
    }

    @Test
    void wrongPasswordProbeRejectsDeniedStream() {
        var input = new BufferedInputStream(new ByteArrayInputStream(
                "wrong password. access denied.".getBytes(StandardCharsets.UTF_8)));
        assertThrows(Exception.class, () -> Http.failOnWrongPassword(input));
    }

    @Test
    void closesStreamAfterHandlerFailureAndMasksError() {
        response("/stream", 200, "payload");
        var stream = new AtomicReference<InputStream>();
        Exception failure = assertThrows(Exception.class, () -> Http.stream(url + "/stream", 2000, input -> {
            stream.set(input);
            throw new Exception("failed password=private-value");
        }));
        assertEquals("failed password=***", failure.getMessage());
        assertThrows(Exception.class, () -> stream.get().read());
    }

    @Test
    void readTimeoutTerminatesRequest() {
        server.createContext("/slow", exchange -> {
            try {
                Thread.sleep(250);
            } catch (InterruptedException interrupted) {
                Thread.currentThread().interrupt();
            } finally {
                exchange.close();
            }
        });
        server.start();
        assertTimeoutPreemptively(java.time.Duration.ofSeconds(3), () ->
                assertThrows(Exception.class, () -> Http.get(url + "/slow", 50)));
    }
}
