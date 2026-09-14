"""Stand-ins for a third-party customer: a public origin hosting their video, and
their webhook receiver."""
import http.server, socketserver, threading, json, os, hmac, hashlib, sys

W = "/private/tmp/claude-501/-Users-ebnsina-Sites-alchemy-stream/5f3fb89c-8e1a-439c-aa7d-1fc7162d5944/scratchpad"
SRC = os.path.join(W, "lecture.mp4")

class Src(http.server.BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path.startswith("/redirect-internal"):
            self.send_response(302)
            self.send_header("Location", "http://169.254.169.254/latest/meta-data/")
            self.end_headers(); return
        self.send_response(200)
        self.send_header("Content-Length", str(os.path.getsize(SRC)))
        self.send_header("Content-Type", "video/mp4")
        self.end_headers()
        with open(SRC, "rb") as f:
            self.wfile.write(f.read())
    def log_message(self, *a): pass

class Hook(http.server.BaseHTTPRequestHandler):
    def do_POST(self):
        n = int(self.headers.get("Content-Length", 0))
        body = self.rfile.read(n)
        with open("/tmp/alctest/hooks.jsonl", "a") as f:
            f.write(json.dumps({
                "sig": self.headers.get("X-Alchemist-Signature"),
                "body": body.decode(),
            }) + "\n")
        self.send_response(200); self.end_headers()
    def log_message(self, *a): pass

socketserver.TCPServer.allow_reuse_address = True
threading.Thread(
    target=lambda: socketserver.TCPServer(("127.0.0.1", 8071), Src).serve_forever(),
    daemon=True).start()
socketserver.TCPServer(("127.0.0.1", 8072), Hook).serve_forever()
