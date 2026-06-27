import json,urllib.request
B="http://127.0.0.1:8080/warm-nest"
def c(p,b,t=None):
    r=urllib.request.Request(B+p,data=json.dumps(b).encode(),method="POST");r.add_header("Content-Type","application/json")
    if t:r.add_header("X-API-Token",t)
    try:
        with urllib.request.urlopen(r,timeout=10) as x:return json.loads(x.read())
    except urllib.error.HTTPError as e:return json.loads(e.read())
def status(t,cid):
    for x in c("/v1/reward/list",{},t)["data"]:
        if x["claimId"]==cid:return x["status"]
u=c("/v1/user/login",{"code":"sm3"})["data"];t=u["token"]
c("/v1/checkin/do",{"photoUrl":"http://x/p","weather":"晴","city":"广安"},t)
cid=c("/v1/reward/list",{},t)["data"][0]["claimId"]
print("初始 status=%s"%status(t,cid))
# machine 状态保护测试：PENDING 状态直接 sign 接口(发approve)→走PENDING的transition(领取)
print("\n--- 验证 machine 按当前状态路由（非接口名）---")
r=c("/admin/v1/reward-claim/sign",{"taskId":cid},t)
print("PENDING调sign接口: result=%s → status=%s（machine按PENDING走了领取transition）"%(r.get("result"),status(t,cid)))
# 正常推进到 SIGNED
c("/admin/v1/reward-claim/ship",{"taskId":cid,"params":{"expressCompany":"SF","expressNo":"S1"}},t)
c("/admin/v1/reward-claim/sign",{"taskId":cid},t)
print("推进到: status=%s"%status(t,cid))
# 关键：终态SIGNED再操作，machine应拒
r=c("/admin/v1/reward-claim/ship",{"taskId":cid,"params":{"expressCompany":"X","expressNo":"Y"}},t)
print("\n终态SIGNED再ship: result=%s msg=%s（machine终态保护）"%(r.get("result"),r.get("message","")[:60]))
