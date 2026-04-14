# 🚀 YourTool

**Programmable tunneling + AI-powered debugging platform (ngrok alternative)**

Expose your localhost, debug traffic, replay requests, and resolve issues instantly.

---

## 📌 Overview

YourTool helps developers:

- Expose local servers to the internet
- Debug and inspect HTTP traffic
- Replay and simulate requests
- Share debugging sessions
- Automatically detect & fix issues using AI

---

## ⚡ Quick Start

### CLI
```bash
npx yourtool start
```

### Node.js
```js
const tunnel = await startTunnel({
  port: 3000,
});

console.log(tunnel.url);

await tunnel.stop();
```

---

## 🧩 Features

---

### 🌍 Public Tunnel

```js
startTunnel({ port: 3000 });
```

Expose localhost instantly via public URL.

---

### 🔗 Stable Subdomain

```js
startTunnel({
  port: 3000,
  subdomain: "my-api",
});
```

- Persistent URL across restarts  
- Useful for webhooks  

---

### 📊 Request Logs

```js
startTunnel({
  port: 3000,
  logs: true,
});
```

- View all requests in real-time  
- Debug easily  

---

### 🔁 Request Replay

```js
await tunnel.replay(requestId);
```

- Retry failed requests  
- No need to trigger external services again  

---

### 🧠 Debug Mode

```js
startTunnel({
  port: 3000,
  mode: "debug",
});
```

Shows:
- Headers  
- Body  
- Status  
- Latency  

---

### 🔌 Hooks

```js
startTunnel({
  port: 3000,
  onRequest: (req) => {},
  onResponse: (res) => {},
});
```

- Inject custom logic  
- Real-time processing  

---

### ✏️ Request / Header Modification

```js
startTunnel({
  port: 3000,
  modifyRequest: (req) => {
    req.headers["x-test"] = "123";
    return req;
  },
});
```

- Simulate edge cases  
- Modify traffic dynamically  

---

### 🧪 Mock Responses

```js
startTunnel({
  port: 3000,
  mock: (req) => {
    if (req.url === "/test") {
      return { status: 200, body: "mocked" };
    }
  },
});
```

- Test without backend  
- Faster development  

---

### 🔄 Modes

```js
mode: "dev"
mode: "debug"
mode: "webhook"
```

---

### 📡 Stats

```js
tunnel.stats();
```

- Requests/sec  
- Errors  
- Latency  

---

### ⏱️ Timing

- Total request time  
- Backend time  
- Network delay  

---

### 🔍 Auto Port Detection

```bash
yourtool start
```

Automatically detects running apps.

---

### 🧾 Structured Logs

```js
logs: {
  format: "json",
  saveToFile: true,
}
```

---

### 🔐 Auth

```js
startTunnel({
  port: 3000,
  auth: true,
});
```

Secure your tunnel.

---

## 🤝 Session Sharing

```js
const session = await tunnel.share();
console.log(session.url);
```

Example:
```
https://yourtool.dev/session/abc123
```

### Why it's useful

- Share debugging session via link  
- No screen sharing needed  
- Collaborate in real-time  

### Use cases

- Team debugging  
- Bug reporting  
- Client demos  

---

## 🤖 AI Debugger & Issue Resolver

```js
startTunnel({
  port: 3000,
  aiDebug: true,
});
```

### What it does

- Detects failed requests  
- Explains errors  
- Suggests fixes  
- Generates example payloads  

### Example

```
❌ 500 Error

AI Suggestion:
Missing field "user_id"

Fix:
{
  "user_id": "123",
  "name": "test"
}
```

---

## 🚀 Why YourTool

| Feature | YourTool | ngrok |
|--------|--------|------|
| Request Replay | ✅ | ❌ |
| Session Sharing | ✅ | ❌ |
| AI Debugging | ✅ | ❌ |
| Hooks | ✅ | ❌ |
| Mocking | ✅ | ❌ |

---

## 📦 CLI

```bash
yourtool start --port 3000
```

---

## 🎯 Vision

A platform for:
- Debugging APIs  
- Testing integrations  
- Collaborating in real-time  

---

## 📣 Contributing

Contributions are welcome.