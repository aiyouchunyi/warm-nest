import json, urllib.request, uuid
BASE="http://127.0.0.1:8080/warm-nest/v1"
PASS=[]; FAIL=[]
def call(path, body, token=None, raw=False):
    req=urllib.request.Request(BASE+path, data=json.dumps(body).encode(), method="POST")
    req.add_header("Content-Type","application/json")
    if token: req.add_header("X-API-Token", token)
    with urllib.request.urlopen(req, timeout=10) as r: return json.loads(r.read())
def chk(name, cond, detail=""):
    (PASS if cond else FAIL).append(name)
    print(("✅" if cond else "❌")+" %-22s %s"%(name,detail))

# 1 登录 A/B
a=call("/user/login",{"code":"child_a"})["data"]; chk("user/login(A)", bool(a.get("token")), "userId=%s"%a["userId"])
b=call("/user/login",{"code":"elder_b"})["data"]; chk("user/login(B)", bool(b.get("token")))
# 2 发邀请
inv=call("/invitation/create",{"elderPhone":"138","relation":"GRANDPA","remindTime":"09:00"},a["token"])["data"]
chk("invitation/create", bool(inv.get("inviteCode")), "code=%s"%inv["inviteCode"])
# 4 邀请列表(之前没测)
il=call("/invitation/list",{},a["token"])
chk("invitation/list", isinstance(il.get("data"),list) and len(il["data"])>=1, "条数=%s"%len(il.get("data") or []))
# 3 接受
acc=call("/invitation/accept",{"inviteCode":inv["inviteCode"]},b["token"])
chk("invitation/accept", acc.get("data",{}).get("ok")==True)
# 5 打卡
ci=call("/checkin/do",{"photoUrl":"http://x/p.png","weather":"晴","city":"广安"},b["token"])["data"]
chk("checkin/do", ci.get("elderUserId")==b["userId"], "elderUserId对=%s"%(ci.get("elderUserId")==b["userId"]))
# 6 今日状态(之前没测)
td=call("/checkin/today",{},b["token"])
chk("checkin/today", td.get("data",{}).get("checked")==True, "checked=%s"%td.get("data",{}).get("checked"))
# 7 月份记录(之前没测)
mo=call("/checkin/month",{"yearMonth":"2026-06"},b["token"])
chk("checkin/month", isinstance(mo.get("data"),list) and len(mo["data"])>=1, "条数=%s"%len(mo.get("data") or []))
# 8 消息列表
msg=call("/message/list",{"type":""},a["token"])["data"]
chk("message/list", msg.get("unread")>=1, "unread=%s"%msg.get("unread"))
# 11 家庭视图
fa=call("/family/view",{},a["token"])["data"]; fb=call("/family/view",{},b["token"])["data"]
chk("family/view(双端)", fa[0]["viewerRole"]=="GUARDIAN" and fb[0]["viewerRole"]=="ELDER", "A=%s B=%s"%(fa[0]["viewerRole"],fb[0]["viewerRole"]))
# 12 改提醒时间(之前没测)
rm=call("/family/remind",{"elderUserId":b["userId"],"remindTime":"10:00"},b["token"])
chk("family/remind", rm.get("data",{}).get("ok")==True)
# 9 奖励列表
rw=call("/reward/list",{},b["token"])
chk("reward/list", isinstance(rw.get("data"),list), "条数=%s"%len(rw.get("data") or []))

print("\n=== 汇总 PASS=%d FAIL=%d ==="%(len(PASS),len(FAIL)))
if FAIL: print("失败:",FAIL)
