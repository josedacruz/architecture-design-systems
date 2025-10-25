from flask import Flask, request, jsonify, make_response, abort
import hashlib, json
from copy import deepcopy

def create_app():
    app = Flask(__name__)

    # --- In-memory "DB"
    DB = {"1": {"title": "Hello", "content": "ETag demo", "version": 1}}

    def canonical_json(data) -> bytes:
        return json.dumps(data, sort_keys=True, separators=(",", ":")).encode()

    def make_etag(doc) -> str:
        body = deepcopy(doc); body.pop("version", None)
        short_hash = hashlib.sha1(canonical_json(body)).hexdigest()[:8]
        return f'W/"{doc["version"]}-{short_hash}"'

    @app.route("/docs/<doc_id>", methods=["GET"])
    def get_doc(doc_id):
        doc = DB.get(doc_id) or abort(404)
        etag = make_etag(doc)
        inm = request.headers.get("If-None-Match")
        if inm and inm == etag:
            resp = make_response("", 304); resp.headers["ETag"] = etag; return resp
        resp = jsonify({k: v for k, v in doc.items() if k != "version"})
        resp.headers["ETag"] = etag
        return resp

    @app.route("/docs/<doc_id>", methods=["PUT"])
    def put_doc(doc_id):
        if_match = request.headers.get("If-Match")
        doc = DB.get(doc_id) or abort(404)
        current = make_etag(doc)
        if if_match and if_match != current:
            return make_response(("Precondition Failed", 412))
        payload = request.get_json(force=True) or {}
        for k in ["title", "content"]:
            if k in payload: doc[k] = payload[k]
        doc["version"] += 1
        new_etag = make_etag(doc)
        resp = jsonify({k: v for k, v in doc.items() if k != "version"})
        resp.headers["ETag"] = new_etag
        return resp, 200

    return app

def main():
    app = create_app()
    # disable the Flask reloader to avoid double-running main()
    app.run(port=5000, debug=True, use_reloader=False)

if __name__ == "__main__":
    main()
