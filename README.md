
IPFS-Based Identity Management API

This is a simple identity management API written in Go, designed to securely store user credentials on the [IPFS](https://ipfs.io/) decentralized network. It supports user creation, authentication, updating, and deletion.

 Features

- 🚀 RESTful API for user management
- 🔐 Secure password hashing with bcrypt
- 📦 User data stored in IPFS
- 🌍 Environment-based configuration
- 🧾 Logging with custom logger

---

 🛠️ Prerequisites

- [Go](https://golang.org/dl/)
- [IPFS Node](https://docs.ipfs.tech/install/)
- [Docker (Optional)](https://www.docker.com/) if running IPFS locally
- IPFS must be running and publicly accessible if not running locally

---

 📂 Project Structure

```
.
├── main.go                 # Main application
├── logger/                 # Custom logger
│   └── logger.go
├── go.mod / go.sum         # Go module files
├── .env                    # Environment variables
└── README.md               # This file
```

---

## 📦 Installation & Run

### 1. Clone the repository

```bash
git clone https://github.com/Umayanga12/IPFS-Blockchain-db.git
```

### 2. Create `.env` file

```ini
# .env
IPFS_NODE=http://localhost:5001
SERVER_ADDR=localhost:8080
```

> Replace `IPFS_NODE` with your local or public IPFS API address.

### 3. Build and run the server

```bash
go mod tidy
go run main.go
```

---

## 📬 API Endpoints

| Method | Endpoint         | Description           |
|--------|------------------|-----------------------|
| POST   | `/users`         | Register new user     |
| POST   | `/login`         | Authenticate user     |
| PUT    | `/users/{id}`    | Update user details   |
| DELETE | `/users/{id}`    | Delete user           |
| GET    | `/`              | Welcome message       |

---

## 🧪 Example Request/Response

### ➕ Register a User

**POST** `/users`

```json
{
  "username": "alice",
  "password": "supersecret"
}
```

**Response**
```json
{
  "id": "9a38f6d6-5e5a-4d2f-b9f3-1c781a0cfb9f"
}
```

---

### 🔑 Login

**POST** `/login`

```json
{
  "username": "alice",
  "password": "supersecret"
}
```

**Response**
```json
{
  "id": "9a38f6d6-5e5a-4d2f-b9f3-1c781a0cfb9f"
}
```

---

### ✏️ Update User

**PUT** `/users/{id}`

```json
{
  "username": "alice_new",
  "password": "newpass"
}
```

**Response:** HTTP 204 No Content

---

### ❌ Delete User

**DELETE** `/users/{id}`

**Response:** HTTP 204 No Content

---

## 📚 Notes

- IPFS stores the latest state by generating a new CID. The backend tracks only the latest CID in memory (`IdentityManager.cid`). You can persist it if needed.
- For production, consider encrypting data and securely storing IPFS CIDs.

