import json, urllib.request, io
BASE="http://127.0.0.1:8080/warm-nest/v1"
def login(code):
    req=urllib.request.Request(BASE+"/user/login",data=json.dumps({"code":code}).encode(),method="POST")
    req.add_header("Content-Type","application/json")
    with urllib.request.urlopen(req,timeout=10) as r: return json.loads(r.read())["data"]["token"]
tok=login("elder_b")
# 构造 multipart 上传一个最小 PNG
png=bytes([0x89,0x50,0x4E,0x47,0x0D,0x0A,0x1A,0x0A,0,0,0,0x0D,0x49,0x48,0x44,0x52,0,0,0,1,0,0,0,1,8,6,0,0,0,0x1F,0x15,0xC4,0x89,0,0,0,0x0A,0x49,0x44,0x41,0x54,0x78,0x9C,0x63,0,1,0,0,5,0,1,0x0D,0x0A,0x2D,0xB4,0,0,0,0,0x49,0x45,0x4E,0x44,0xAE,0x42,0x60,0x82])
boundary="----wnboundary"
body=io.BytesIO()
body.write(("--%s\r\nContent-Disposition: form-data; name=\"file\"; filename=\"t.png\"\r\nContent-Type: image/png\r\n\r\n"%boundary).encode())
body.write(png); body.write(("\r\n--%s--\r\n"%boundary).encode())
req=urllib.request.Request(BASE+"/upload",data=body.getvalue(),method="POST")
req.add_header("Content-Type","multipart/form-data; boundary=%s"%boundary)
req.add_header("X-API-Token",tok)
with urllib.request.urlopen(req,timeout=10) as r: resp=json.loads(r.read())
print("[upload]", json.dumps(resp,ensure_ascii=False))
url=resp.get("data",{}).get("url","")
print("✅ upload" if url else "❌ upload")
# static 读取：用返回的 url 直接 GET
if url:
    with urllib.request.urlopen(url,timeout=10) as r:
        data=r.read()
    print("✅ static 读取 字节数=%d (PNG头=%s)"%(len(data), data[:4]==bytes([0x89,0x50,0x4E,0x47])))
