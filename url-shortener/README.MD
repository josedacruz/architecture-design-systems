# 🔗 URL Shortener (Go)

A lightweight, in-memory URL shortener built in **Go**, designed for simplicity, clarity, and educational use. It allows users to shorten long URLs using a simple Base62-encoded ID and redirect using the short code.

---

## 📦 Features

- Shorten long URLs into Base62 codes (e.g., `http://localhost:8080/1`)
- Redirect users from short URLs back to the original
- In-memory storage (no persistence; resets on restart)
- Built-in HTTP server with clean endpoints
- Self-contained, no third-party dependencies

---

## 🛠️ Getting Started

### 🔁 Clone the repository

```bash
git clone https://github.com/josedacruz/architecture-design-systems.git
cd url-shortener
````

> Make sure your Go environment is set up: [https://go.dev/doc/install](https://go.dev/doc/install)

---

### 🧱 Build and Run Locally

```bash
cd cmd/shortener
go run main.go
```

By default, the server runs at:
`http://localhost:8080`

You can override settings using environment variables:

```bash
PORT=9090 BASE_URL="http://localhost:9090/" go run main.go
```

---

## 🔬 Testing the API

You can use `curl`, Postman, or any REST client.

### 🔗 Shorten a URL

**POST** `http://localhost:8080/shorten`
**Headers:**

```
Content-Type: application/json
```

**Body:**

```json
{
  "long_url": "http://www.reddit.com"
}
```

**Response:**

```json
{
  "short_url": "http://localhost:8080/1"
}
```

---

### 🚀 Redirect to Original URL

**GET** `http://localhost:8080/1`

This will redirect (HTTP 301) to:

```
http://www.reddit.com
```

---

## 📁 Project Structure

```
url-shortener/
├── cmd/shortener/         # Main application entrypoint
├── internal/
│   ├── handler/           # HTTP handlers (shorten, redirect)
│   ├── service/           # Business logic
│   ├── storage/           # In-memory storage & ID management
│   └── model/             # Request/response types
├── pkg/                   # Shared logic (e.g., base62, urls)
└── go.mod
```

---

## 🧠 Future Ideas

* Add persistent storage (e.g., Redis, SQLite)
* Custom short codes
* Expiry logic for short URLs
* Metrics and logging middleware
* Central generation of ids (if multiple servers)

---

## ⚖️ License

MIT — feel free to use, improve, and contribute!

## Support me!

Please visit and subscribe: [ITArchitecture4All youtube channel](https://www.youtube.com/@itarchitecture4all)

