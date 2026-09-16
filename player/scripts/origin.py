#!/usr/bin/env python3
"""Local stand-in for the Alchemist origin, for developing the player offline.

It mirrors exactly the parts of `internal/modules/delivery` the player touches:
the same HMAC over `prefix|exp` truncated to 32 hex chars, the same manifest
rewriting so relative URIs stay authorized, the same raw 16-byte content key, the
same always-204 beacon. Point the player at a real Alchemist and nothing changes.

  python3 scripts/origin.py            # serves ./fixture on :8099
  python3 scripts/origin.py --sign     # prints a signed master.m3u8 URL and exits
"""
import argparse, hashlib, hmac, json, mimetypes, os, re, sys, time
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from urllib.parse import urlparse, parse_qs

ROOT = os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "fixture")
KEYRING = {"k1": "dev-only-playback-key-change-me-32b"}
CONTENT_KEY = bytes.fromhex("166634c675823c356a6a1ea56f4da1cd")
SIG_LEN = 32

HLS_ATTR_URI = re.compile(r'URI="([^"]+)"')
DASH_BASE = re.compile(r"<BaseURL>([^<]+)</BaseURL>")


def compute(secret: str, prefix: str, exp: str) -> str:
    return hmac.new(secret.encode(), f"{prefix}|{exp}".encode(), hashlib.sha256).hexdigest()[:SIG_LEN]


def verify(prefix: str, kid: str, sig: str, exp: str) -> bool:
    secret = KEYRING.get(kid)
    if not secret or not sig or not exp:
        return False
    try:
        if int(exp) < time.time():
            return False
    except ValueError:
        return False
    return hmac.compare_digest(compute(secret, prefix, exp), sig)


def append_query(uri: str, query: str) -> str:
    if "://" in uri or uri.startswith("//"):
        return uri
    return uri + ("&" if "?" in uri else "?") + query


def sign_manifest(body: bytes, query: str, name: str) -> bytes:
    if not query:
        return body
    if name.endswith(".mpd"):
        q = query.replace("&", "&amp;")
        return DASH_BASE.sub(lambda m: f"<BaseURL>{append_query(m.group(1), q)}</BaseURL>", body.decode()).encode()
    out = []
    for line in body.decode().split("\n"):
        s = line.strip()
        if not s:
            out.append(line)
        elif s.startswith("#"):
            out.append(HLS_ATTR_URI.sub(lambda m: 'URI="%s"' % append_query(m.group(1), query), line))
        else:
            out.append(append_query(s, query))
    return "\n".join(out).encode()


class Handler(BaseHTTPRequestHandler):
    protocol_version = "HTTP/1.1"

    def log_message(self, fmt, *a):
        sys.stderr.write("  %s\n" % (fmt % a))

    def cors(self):
        self.send_header("Access-Control-Allow-Origin", "*")
        self.send_header("Access-Control-Allow-Methods", "GET, HEAD, POST, OPTIONS")
        self.send_header("Access-Control-Allow-Headers", "Range, If-None-Match, Content-Type")
        self.send_header("Access-Control-Expose-Headers", "Content-Length, Content-Range, ETag, Accept-Ranges")

    def parts(self):
        u = urlparse(self.path)
        seg = [s for s in u.path.split("/") if s]
        if len(seg) < 4 or seg[0] != "playback":
            return None
        q = parse_qs(u.query)
        one = lambda k: (q.get(k) or [""])[0]
        return {
            "tenant": seg[1], "asset": seg[2], "file": seg[3],
            "prefix": "/playback/%s/%s" % (seg[1], seg[2]),
            "query": u.query, "kid": one("kid"), "sig": one("sig"), "exp": one("exp"),
        }

    def refuse(self, code, body=b""):
        self.send_response(code)
        self.cors()
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        if body and self.command != "HEAD":
            self.wfile.write(body)

    def do_OPTIONS(self):
        self.refuse(204)

    def do_POST(self):
        p = self.parts()
        if not p or p["file"] != "beacon":
            return self.refuse(404)
        if not verify(p["prefix"], p["kid"], p["sig"], p["exp"]):
            return self.refuse(403)
        n = int(self.headers.get("Content-Length") or 0)
        raw = self.rfile.read(n) if n else b""
        try:
            sys.stderr.write("  beacon %s\n" % json.dumps(json.loads(raw or b"{}")))
        except Exception:
            sys.stderr.write("  beacon (unparseable, still 204)\n")
        self.refuse(204)  # always 204, like the real one

    def do_HEAD(self):
        self.do_GET()

    def do_GET(self):
        p = self.parts()
        if not p:
            return self.refuse(404, b"not a playback path")
        if not verify(p["prefix"], p["kid"], p["sig"], p["exp"]):
            return self.refuse(403, b'{"error":{"code":"playback_not_authorized"}}')

        if p["file"] == "key":
            self.send_response(200)
            self.cors()
            self.send_header("Content-Type", "application/octet-stream")
            self.send_header("Cache-Control", "no-store")
            self.send_header("Content-Length", str(len(CONTENT_KEY)))
            self.end_headers()
            if self.command != "HEAD":
                self.wfile.write(CONTENT_KEY)
            return

        path = os.path.join(ROOT, "cmaf", p["tenant"], p["asset"], os.path.basename(p["file"]))
        if not os.path.isfile(path):
            return self.refuse(404, b'{"error":{"code":"not_found"}}')

        manifest = p["file"].endswith((".m3u8", ".mpd"))
        with open(path, "rb") as f:
            data = f.read()
        if manifest:
            data = sign_manifest(data, p["query"], p["file"])

        ctype = {
            ".m3u8": "application/vnd.apple.mpegurl", ".mpd": "application/dash+xml",
            ".cmfv": "video/mp4", ".cmfa": "audio/mp4", ".vtt": "text/vtt",
        }.get(os.path.splitext(p["file"])[1], mimetypes.guess_type(path)[0] or "application/octet-stream")

        rng = self.headers.get("Range")
        status, start, end = 200, 0, len(data) - 1
        if rng and not manifest:
            m = re.match(r"bytes=(\d*)-(\d*)", rng)
            if m:
                start = int(m.group(1) or 0)
                end = int(m.group(2)) if m.group(2) else len(data) - 1
                if start >= len(data):
                    return self.refuse(416)
                end = min(end, len(data) - 1)
                status = 206

        body = data[start:end + 1]
        self.send_response(status)
        self.cors()
        self.send_header("Content-Type", ctype)
        self.send_header("Accept-Ranges", "bytes")
        self.send_header("Cache-Control", "public, max-age=2" if manifest else "public, max-age=31536000, immutable")
        if status == 206:
            self.send_header("Content-Range", "bytes %d-%d/%d" % (start, end, len(data)))
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        if self.command != "HEAD":
            self.wfile.write(body)


def sign_url(origin, tenant, asset, file, ttl):
    exp = str(int(time.time()) + ttl)
    prefix = "/playback/%s/%s" % (tenant, asset)
    sig = compute(KEYRING["k1"], prefix, exp)
    return "%s%s/%s?exp=%s&kid=k1&sig=%s" % (origin, prefix, file, exp, sig)


if __name__ == "__main__":
    ap = argparse.ArgumentParser()
    ap.add_argument("--port", type=int, default=8099)
    ap.add_argument("--origin", default="http://localhost:8099")
    ap.add_argument("--tenant", default="t_dev")
    ap.add_argument("--asset", default="a_dev")
    ap.add_argument("--ttl", type=int, default=4 * 3600)
    ap.add_argument("--sign", action="store_true", help="print a signed URL and exit")
    a = ap.parse_args()

    if a.sign:
        print(sign_url(a.origin, a.tenant, a.asset, "master.m3u8", a.ttl))
        raise SystemExit(0)

    print("origin stand-in on http://localhost:%d serving %s" % (a.port, os.path.normpath(ROOT)))
    print("signed URL:\n  %s\n" % sign_url(a.origin, a.tenant, a.asset, "master.m3u8", a.ttl))
    ThreadingHTTPServer(("127.0.0.1", a.port), Handler).serve_forever()
