import http2 from "http2";
import WebSocket from "ws";
import fs from "fs";
import extractJsonFromString from "extract-json-from-string";

const client = http2.connect('https://canary.discord.com');

let vanity;
let mfaToken = "";
const guilds = {};
const token = ""; // self hesabinizin tokeni 

const log = (data) => {
  const ext = extractJsonFromString(data.toString());
  const find = ext.find((e) => e.code || e.message);
  if (find) {
    const body = JSON.stringify({
      content: `@everyone ${vanity}\n\`\`\`json\n${JSON.stringify(find)}\`\`\``
    });
    const req = client.request({
      ':method': 'POST',
      ':path': '/api/channels/CHANNEL İD YAZACAKSİNİZ/messages',
      'authorization': token,
      'content-type': 'application/json',
    });
    req.write(body);
    req.end();
  }
};

client.on('connect', async () => {
  connectWebSocket();
});

client.on('error', () => {
  process.exit();
});

client.on('close', () => {
  process.exit();
});

const connectWebSocket = () => {
  const ws = new WebSocket("wss://gateway.discord.gg/");
  ws.onclose = () => process.exit();
  ws.onmessage = (message) => {
    const { d, op, t } = JSON.parse(message.data);
    switch (t) {
      case "GUILD_UPDATE": {
        const find = guilds[d.guild_id];
        if (find && find !== d.vanity_url_code) {
          vanity = find;
          const requestBody = JSON.stringify({ code: find });
          const req = client.request({
            ':method': 'PATCH',
            ':path': '/api/v9/guilds/SUNUCU İD YAZACAKSINIZ/vanity-url',
            'authorization': token,
            'x-discord-mfa-authorization': mfaToken,
            'user-agent': 'Chrome/124',
            'x-super-properties': 'eyJicm93c2VyIjoiQ2hyb21lIiwiYnJvd3Nlcl91c2VyX2FnZW50IjoiQ2hyb21lIiwiY2xpZW50X2J1aWxkX251bWJlciI6MzU1NjI0fQ==',
            'content-type': 'application/json',
          }, { priority: { weight: 255, exclusive: true } });
          let responseData = '';
          req.on("data", chunk => responseData += chunk);
          req.on("end", () => {
            log(responseData);
            vanity = find;
          });
          req.end(requestBody);
        }
        break;
      }
      case "READY": {
        d.guilds.forEach(({ id, vanity_url_code }) => {
          if (vanity_url_code) guilds[id] = vanity_url_code;
        });
        break;
      }
    }
    if (op === 7) return process.exit();
    if (op === 10) {
      ws.send(JSON.stringify({
        op: 2,
        d: {
          token: token,
          intents: 1 << 0,
          properties: {
            os: "linux",
            browser: "firefox",
            device: "1337",
          },
        },
      }));
      setTimeout(() => {
        setInterval(() => ws.send(JSON.stringify({ op: 1, d: {}, s: null, t: "heartbeat" })), d.heartbeat_interval);
      }, d.heartbeat_interval * Math.random());
    }
  };
};

setInterval(() => {
  if (client.destroyed) process.exit();
  const req = client.request({
    ':method': 'HEAD',
    ':path': '/api/users/@me',
    'authorization': token,
  });
  req.on('error', () => { });
  req.end();
}, 2000);

fs.readFile("mfa_token.txt", "utf8", (err, data) => {
  if (!err) mfaToken = data.trim();
});
fs.watch("mfa_token.txt", (eventType) => {
  if (eventType === "change") {
    fs.readFile("mfa_token.txt", "utf8", (err, data) => {
      if (!err) mfaToken = data.trim();
    });
  }
});
