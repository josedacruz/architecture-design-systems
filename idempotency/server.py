from flask import Flask, request, jsonify

app = Flask(__name__)
processed = {}  # store idempotency results

@app.route("/payments", methods=["POST"])
def create_payment():
    key = request.headers.get("Idempotency-Key")
    data = request.get_json(force=True)
    if not key:
        return {"error": "Missing Idempotency-Key"}, 400
    if key in processed:
        return processed[key]  # same response as before
    payment = {"id": len(processed) + 1, "amount": data["amount"]}
    processed[key] = (payment, 200)
    return payment, 200

if __name__ == "__main__":
    app.run(port=5000)