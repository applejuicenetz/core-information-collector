package de.applejuicenet.collector;

public class Target {

    private final String url;
    private final String token;
    private final String line;

    public Target(String url, String token, String line) {
        this.url = url;
        this.token = token;
        this.line = line;
    }

    public String getUrl() {
        return url;
    }

    public String getToken() {
        return token;
    }

    public String getLine() {
        return line;
    }
}

