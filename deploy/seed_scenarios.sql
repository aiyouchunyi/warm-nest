-- warm-nest 全场景测试数据（SQL 一键清空 + 重建，供前端测试）
-- 用法（宿主机）：mysql -h127.0.0.1 -u<user> -p<pwd> warm-nest < deploy/seed_scenarios.sql
-- 幂等可重复跑：先清空全部业务表再重建。日期用 CURDATE() 动态算，任何时候跑都对齐当天。
--
-- 账号登录方式（mock）：前端 POST /warm-nest/v1/user/login {"code":"<code>"}，
--   code=test_xxx → 命中 openid=mock_openid_test_xxx 的预置账号，拿到其 token。
--   userId 已在本脚本固定（见下方账号表），与 openid 绑定。
--
-- 时间口径：打卡日期 Asia/Shanghai，与服务端一致（DB 连接 loc=Local，容器 TZ 已设）。

SET @now := UNIX_TIMESTAMP() * 1000;          -- 当前毫秒
SET @month_start := CONCAT(DATE_FORMAT(CURDATE(), '%Y-%m'), '-01');

-- ========== 1. 清空全部业务表（含账号）==========
DELETE FROM t_check_in        WHERE 1=1;
DELETE FROM t_message         WHERE 1=1;
DELETE FROM t_reward_claim    WHERE 1=1;
DELETE FROM t_reward_task     WHERE 1=1;
DELETE FROM t_guardianship    WHERE 1=1;
DELETE FROM t_invitation      WHERE 1=1;
DELETE FROM t_shipping_address WHERE 1=1;   -- 需求3 地址簿表（领取从此取址，重建必须清+造）
DELETE FROM t_elder_profile   WHERE 1=1;
DELETE FROM t_fan             WHERE 1=1;
DELETE FROM t_user            WHERE 1=1;

-- ========== 2. 奖励规则种子 ==========
INSERT INTO t_reward_task (task_key,name,`desc`,condition_type,condition_value,reward_kind,reward_name,reward_spec,quantity,reward_params,enable,created_at,updated_at) VALUES
 ('monthly_egg','本月打卡领鸡蛋','本月打卡满（当月天数-3）天，领一盒安心鸡蛋','MONTHLY_CHECK_IN',0,'EGG','安心鸡蛋','30枚/盒',1,'{"sku":"EGG-30"}',1,@now,@now),
 ('continuous_egg_7','连续打卡7天领鸡蛋','连续打卡满 7 天，领一盒安心鸡蛋','CONTINUOUS_CHECK_IN',7,'EGG','安心鸡蛋','30枚/盒',1,'{"sku":"EGG-30"}',1,@now,@now),
 ('cumulative_egg_30','累计打卡30天领鸡蛋','累计打卡满 30 天，领一盒安心鸡蛋','CUMULATIVE_CHECK_IN',30,'EGG','安心鸡蛋','30枚/盒',1,'{"sku":"EGG-30"}',1,@now,@now);

-- ========== 3. 账号（user_id 固定，openid=mock_openid_<code>）==========
-- 老人账号（union_id 用于关联服务号关注态 fan）
INSERT INTO t_user (user_id,open_id,union_id,phone,nickname,avatar,status,last_active_role,created_at,updated_at) VALUES
 ('U_ELDER_RICH','mock_openid_test_elder_rich','UNION_ELDER_RICH','','张奶奶(完整)','','NORMAL','ELDER',@now,@now),
 ('U_ELDER_NOADDR','mock_openid_test_elder_noaddr','UNION_ELDER_NOADDR','','李爷爷(缺地址)','','NORMAL','ELDER',@now,@now),
 ('U_ELDER_SIGNED','mock_openid_test_elder_signed','UNION_ELDER_SIGNED','','王奶奶(已签收)','','NORMAL','ELDER',@now,@now),
 ('U_ELDER_NEW','mock_openid_test_elder_new','UNION_ELDER_NEW','','赵爷爷(无打卡)','','NORMAL','ELDER',@now,@now),
 ('U_ELDER_LONELY','mock_openid_test_elder_lonely','UNION_ELDER_LONELY','','刘爷爷(未绑定)','','NORMAL','',@now,@now);
-- 子女账号
INSERT INTO t_user (user_id,open_id,union_id,phone,nickname,avatar,status,last_active_role,created_at,updated_at) VALUES
 ('U_GUARD_RICH','mock_openid_test_guardian_rich','UNION_GUARD_RICH','','小张(子女,守护多人)','','NORMAL','GUARDIAN',@now,@now),
 ('U_GUARD_PENDING','mock_openid_test_guardian_pending','UNION_GUARD_PENDING','','小陈(仅发邀请)','','NORMAL','GUARDIAN',@now,@now);

-- 服务号关注态（fan）：rich 老人已关注、guardian 子女已关注（测 subscribe-status=true）；noaddr 老人未关注记录（=false）
INSERT INTO t_fan (union_id,official_open_id,subscribed,subscribe_at,created_at,updated_at) VALUES
 ('UNION_ELDER_RICH','off_openid_elder_rich',1,@now,@now,@now),
 ('UNION_GUARD_RICH','off_openid_guard_rich',1,@now,@now,@now),
 ('UNION_ELDER_SIGNED','off_openid_elder_signed',0,0,@now,@now);

-- ========== 4. 守护关系（子女 → 老人，ACTIVE）==========
-- 小张守护 4 位老人（测一子女多老人）：rich/noaddr/signed/new
INSERT INTO t_guardianship (guardianship_id,guardian_user_id,elder_user_id,relation,status,activated_at,created_at,updated_at) VALUES
 ('G_RICH','U_GUARD_RICH','U_ELDER_RICH','GRANDMA','ACTIVE',@now,@now,@now),
 ('G_NOADDR','U_GUARD_RICH','U_ELDER_NOADDR','GRANDPA','ACTIVE',@now,@now,@now),
 ('G_SIGNED','U_GUARD_RICH','U_ELDER_SIGNED','GRANDMA','ACTIVE',@now,@now,@now),
 ('G_NEW','U_GUARD_RICH','U_ELDER_NEW','GRANDPA','ACTIVE',@now,@now,@now);
-- U_ELDER_LONELY 无守护关系（测 bind-status 拦截）
-- U_GUARD_PENDING 无守护关系，只有 PENDING 邀请（测邀请待接受）

-- ========== 5. 老人档案（结构化地址）==========
INSERT INTO t_elder_profile (user_id,real_name,city,remind_time,health_note,address,elder_phone,guardian_phone,created_at,updated_at) VALUES
 ('U_ELDER_RICH','张秀英','上海','09:00','高血压，需按时服药',
   '{"province":"上海市","city":"上海市","district":"浦东新区","street":"陆家嘴街道","detail":"世纪大道100号1栋101","receiverName":"小张","receiverPhone":"13800138001"}',
   '13900139001','13800138001',@now,@now),
 -- 缺地址老人：收货人/电话留空（测领取错误码 10002002/10002003）
 ('U_ELDER_NOADDR','李建国','北京','08:30','',
   '{"province":"北京市","city":"北京市","district":"海淀区","street":"","detail":"中关村大街1号","receiverName":"","receiverPhone":""}',
   '13900139002','',@now,@now),
 ('U_ELDER_SIGNED','王翠兰','广州','10:00','',
   '{"province":"广东省","city":"广州市","district":"天河区","street":"天河路街道","detail":"体育西路1号","receiverName":"小王","receiverPhone":"13800138003"}',
   '13900139003','13800138003',@now,@now),
 ('U_ELDER_NEW','赵德发','成都','09:30','',
   '{"province":"四川省","city":"成都市","district":"武侯区","street":"","detail":"人民南路1号","receiverName":"小赵","receiverPhone":"13800138004"}',
   '13900139004','13800138004',@now,@now);
-- U_ELDER_LONELY 无档案（未绑定）

-- ========== 5b. 收货地址簿（需求3：领取从此表取址，每老人一条默认地址）==========
-- 与档案地址保持一致;NOADDR 老人故意造「缺收货人/电话」以保留领取错误码(10002002/10002003)测试用途。
-- address 列为结构化 json，is_default=1 默认地址，sort_no=0。
INSERT INTO t_shipping_address (address_id,elder_user_id,address,is_default,sort_no,created_at,updated_at) VALUES
 ('ADDR_RICH','U_ELDER_RICH',
   '{"province":"上海市","city":"上海市","district":"浦东新区","street":"陆家嘴街道","detail":"世纪大道100号1栋101","receiverName":"小张","receiverPhone":"13800138001"}',
   1,0,@now,@now),
 -- 缺地址老人：收货人/电话留空（测领取被拦 10002002/10002003）
 ('ADDR_NOADDR','U_ELDER_NOADDR',
   '{"province":"北京市","city":"北京市","district":"海淀区","street":"","detail":"中关村大街1号","receiverName":"","receiverPhone":""}',
   1,0,@now,@now),
 ('ADDR_SIGNED','U_ELDER_SIGNED',
   '{"province":"广东省","city":"广州市","district":"天河区","street":"天河路街道","detail":"体育西路1号","receiverName":"小王","receiverPhone":"13800138003"}',
   1,0,@now,@now),
 ('ADDR_NEW','U_ELDER_NEW',
   '{"province":"四川省","city":"成都市","district":"武侯区","street":"","detail":"人民南路1号","receiverName":"小赵","receiverPhone":"13800138004"}',
   1,0,@now,@now);
-- 张奶奶(rich)额外加一条非默认地址（测多地址列表 + set-default 切换）
INSERT INTO t_shipping_address (address_id,elder_user_id,address,is_default,sort_no,created_at,updated_at) VALUES
 ('ADDR_RICH_2','U_ELDER_RICH',
   '{"province":"上海市","city":"上海市","district":"徐汇区","street":"漕河泾街道","detail":"宜山路500号2栋202","receiverName":"小张","receiverPhone":"13800138009"}',
   0,1,@now,@now);

-- ========== 6. 打卡记录 ==========
-- 张奶奶(rich)：昨天往前连续 8 天打卡（今天留空，前端可真实打卡触发消息）
INSERT INTO t_check_in (check_in_id,elder_user_id,check_in_date,kind,photo_url,weather,city,created_at,updated_at) VALUES
 ('CK_R1','U_ELDER_RICH',DATE_FORMAT(DATE_SUB(CURDATE(),INTERVAL 1 DAY),'%Y-%m-%d'),'NORMAL','https://cdn.example.com/checkin/r1.jpg','晴 26°C','上海',@now,@now),
 ('CK_R2','U_ELDER_RICH',DATE_FORMAT(DATE_SUB(CURDATE(),INTERVAL 2 DAY),'%Y-%m-%d'),'NORMAL','https://cdn.example.com/checkin/r2.jpg','多云 24°C','上海',@now,@now),
 ('CK_R3','U_ELDER_RICH',DATE_FORMAT(DATE_SUB(CURDATE(),INTERVAL 3 DAY),'%Y-%m-%d'),'NORMAL','https://cdn.example.com/checkin/r3.jpg','小雨 22°C','上海',@now,@now),
 ('CK_R4','U_ELDER_RICH',DATE_FORMAT(DATE_SUB(CURDATE(),INTERVAL 4 DAY),'%Y-%m-%d'),'NORMAL','https://cdn.example.com/checkin/r4.jpg','晴 25°C','上海',@now,@now),
 ('CK_R5','U_ELDER_RICH',DATE_FORMAT(DATE_SUB(CURDATE(),INTERVAL 5 DAY),'%Y-%m-%d'),'NORMAL','https://cdn.example.com/checkin/r5.jpg','晴 27°C','上海',@now,@now),
 ('CK_R6','U_ELDER_RICH',DATE_FORMAT(DATE_SUB(CURDATE(),INTERVAL 6 DAY),'%Y-%m-%d'),'NORMAL','https://cdn.example.com/checkin/r6.jpg','阴 23°C','上海',@now,@now),
 ('CK_R7','U_ELDER_RICH',DATE_FORMAT(DATE_SUB(CURDATE(),INTERVAL 7 DAY),'%Y-%m-%d'),'NORMAL','https://cdn.example.com/checkin/r7.jpg','晴 26°C','上海',@now,@now),
 ('CK_R8','U_ELDER_RICH',DATE_FORMAT(DATE_SUB(CURDATE(),INTERVAL 8 DAY),'%Y-%m-%d'),'NORMAL','https://cdn.example.com/checkin/r8.jpg','晴 25°C','上海',@now,@now);
-- 缺地址老人：也连续 8 天（有可领记录，但领取会因缺收货人被拦）
INSERT INTO t_check_in (check_in_id,elder_user_id,check_in_date,kind,photo_url,weather,city,created_at,updated_at) VALUES
 ('CK_N1','U_ELDER_NOADDR',DATE_FORMAT(DATE_SUB(CURDATE(),INTERVAL 1 DAY),'%Y-%m-%d'),'NORMAL','https://cdn.example.com/checkin/n1.jpg','晴','北京',@now,@now),
 ('CK_N2','U_ELDER_NOADDR',DATE_FORMAT(DATE_SUB(CURDATE(),INTERVAL 2 DAY),'%Y-%m-%d'),'NORMAL','https://cdn.example.com/checkin/n2.jpg','晴','北京',@now,@now),
 ('CK_N3','U_ELDER_NOADDR',DATE_FORMAT(DATE_SUB(CURDATE(),INTERVAL 3 DAY),'%Y-%m-%d'),'NORMAL','https://cdn.example.com/checkin/n3.jpg','晴','北京',@now,@now),
 ('CK_N4','U_ELDER_NOADDR',DATE_FORMAT(DATE_SUB(CURDATE(),INTERVAL 4 DAY),'%Y-%m-%d'),'NORMAL','https://cdn.example.com/checkin/n4.jpg','晴','北京',@now,@now),
 ('CK_N5','U_ELDER_NOADDR',DATE_FORMAT(DATE_SUB(CURDATE(),INTERVAL 5 DAY),'%Y-%m-%d'),'NORMAL','https://cdn.example.com/checkin/n5.jpg','晴','北京',@now,@now),
 ('CK_N6','U_ELDER_NOADDR',DATE_FORMAT(DATE_SUB(CURDATE(),INTERVAL 6 DAY),'%Y-%m-%d'),'NORMAL','https://cdn.example.com/checkin/n6.jpg','晴','北京',@now,@now),
 ('CK_N7','U_ELDER_NOADDR',DATE_FORMAT(DATE_SUB(CURDATE(),INTERVAL 7 DAY),'%Y-%m-%d'),'NORMAL','https://cdn.example.com/checkin/n7.jpg','晴','北京',@now,@now);
-- 已签收老人：连续打卡 + 已有完整领取（见下）
INSERT INTO t_check_in (check_in_id,elder_user_id,check_in_date,kind,photo_url,weather,city,created_at,updated_at) VALUES
 ('CK_S1','U_ELDER_SIGNED',DATE_FORMAT(DATE_SUB(CURDATE(),INTERVAL 1 DAY),'%Y-%m-%d'),'NORMAL','https://cdn.example.com/checkin/s1.jpg','晴','广州',@now,@now),
 ('CK_S2','U_ELDER_SIGNED',DATE_FORMAT(DATE_SUB(CURDATE(),INTERVAL 2 DAY),'%Y-%m-%d'),'NORMAL','https://cdn.example.com/checkin/s2.jpg','晴','广州',@now,@now),
 ('CK_S3','U_ELDER_SIGNED',DATE_FORMAT(DATE_SUB(CURDATE(),INTERVAL 3 DAY),'%Y-%m-%d'),'NORMAL','https://cdn.example.com/checkin/s3.jpg','晴','广州',@now,@now);
-- U_ELDER_NEW 无打卡（测空状态）

-- ========== 7. 领取记录（覆盖 PENDING/CLAIMED/SHIPPED/SIGNED 全状态；收货快照=结构化 json）==========
SET @period := CONCAT('continuous-', DATE_FORMAT(DATE_SUB(CURDATE(),INTERVAL 1 DAY),'%Y-%m-%d'), '-7');
SET @month_period := CONCAT('monthly-', DATE_FORMAT(CURDATE(),'%Y-%m'));
-- 收货快照 json（已领/发货/签收的单子带，待领取为空对象）
SET @addr_rich := '{"province":"上海市","city":"上海市","district":"浦东新区","street":"陆家嘴街道","detail":"世纪大道100号1栋101","receiverName":"小张","receiverPhone":"13800138001"}';
SET @addr_new := '{"province":"四川省","city":"成都市","district":"武侯区","street":"","detail":"人民南路1号","receiverName":"小赵","receiverPhone":"13800138004"}';
SET @addr_signed := '{"province":"广东省","city":"广州市","district":"天河区","street":"天河路街道","detail":"体育西路1号","receiverName":"小王","receiverPhone":"13800138003"}';
-- 张奶奶(rich)：PENDING 待领取（前端可测领取动作，地址完整→成功）；收货快照待领取时为空 {}
INSERT INTO t_reward_claim (claim_id,elder_user_id,task_key,period_key,reward_kind,reward_name,reward_spec,quantity,reward_params,achieved_snap,express_detail,status_infos,runtime,`lock`,status,receiver_address,created_at,updated_at) VALUES
 ('CLAIM_PENDING','U_ELDER_RICH','continuous_egg_7',@period,'EGG','安心鸡蛋','30枚/盒',1,'{"sku":"EGG-30"}','{"continuousDays":8}','{}','[]','{}',0,'PENDING','{}',@now,@now);
-- 缺地址老人：PENDING（前端测领取会被拦 10002002 缺收货人）
INSERT INTO t_reward_claim (claim_id,elder_user_id,task_key,period_key,reward_kind,reward_name,reward_spec,quantity,reward_params,achieved_snap,express_detail,status_infos,runtime,`lock`,status,receiver_address,created_at,updated_at) VALUES
 ('CLAIM_NOADDR','U_ELDER_NOADDR','continuous_egg_7',@period,'EGG','安心鸡蛋','30枚/盒',1,'{"sku":"EGG-30"}','{"continuousDays":7}','{}','[]','{}',0,'PENDING','{}',@now,@now);
-- 张奶奶(rich) 月度达标 PENDING（测月度奖励，月概览显示该月 CLAIMABLE）
INSERT INTO t_reward_claim (claim_id,elder_user_id,task_key,period_key,reward_kind,reward_name,reward_spec,quantity,reward_params,achieved_snap,express_detail,status_infos,runtime,`lock`,status,receiver_address,created_at,updated_at) VALUES
 ('CLAIM_MONTHLY','U_ELDER_RICH','monthly_egg',@month_period,'EGG','安心鸡蛋','30枚/盒',1,'{"sku":"EGG-30"}','{"month":"本月","days":8,"threshold":27}','{}','[]','{}',0,'PENDING','{}',@now,@now);
-- 张奶奶(rich) 累计达标 CLAIMED 已领取待发货（收货快照=结构化 json）
INSERT INTO t_reward_claim (claim_id,elder_user_id,task_key,period_key,reward_kind,reward_name,reward_spec,quantity,reward_params,achieved_snap,express_detail,status_infos,runtime,`lock`,status,receiver_address,claimed_at,created_at,updated_at) VALUES
 ('CLAIM_CLAIMED','U_ELDER_RICH','cumulative_egg_30','cumulative-30','EGG','安心鸡蛋','30枚/盒',1,'{"sku":"EGG-30"}','{"cumulativeDays":30}','{}','[]','{}',0,'CLAIMED',@addr_rich,@now,@now,@now);
-- 赵爷爷(new) 已发货 SHIPPED（有快递无签收）
INSERT INTO t_reward_claim (claim_id,elder_user_id,task_key,period_key,reward_kind,reward_name,reward_spec,quantity,reward_params,achieved_snap,express_company,express_no,express_detail,status_infos,runtime,`lock`,status,receiver_address,claimed_at,shipped_at,created_at,updated_at) VALUES
 ('CLAIM_SHIPPED','U_ELDER_NEW','cumulative_egg_30','cumulative-30','EGG','安心鸡蛋','30枚/盒',1,'{"sku":"EGG-30"}','{"cumulativeDays":30}','圆通速递','YT9876543210','{}','[]','{}',0,'SHIPPED',@addr_new,@now,@now,@now,@now);
-- 王奶奶(signed)：已签收终态（测领取全字段+物流轨迹展示）
INSERT INTO t_reward_claim (claim_id,elder_user_id,task_key,period_key,reward_kind,reward_name,reward_spec,quantity,reward_params,achieved_snap,express_company,express_no,express_detail,status_infos,runtime,`lock`,status,receiver_address,claimed_at,shipped_at,signed_at,created_at,updated_at) VALUES
 ('CLAIM_SIGNED','U_ELDER_SIGNED','continuous_egg_7',@period,'EGG','安心鸡蛋','30枚/盒',1,'{"sku":"EGG-30"}','{"continuousDays":3}','顺丰速运','SF1234567890','{}','[]','{}',0,'SIGNED',@addr_signed,@now,@now,@now,@now,@now);

-- ========== 8. 消息（子女端打卡通知，带照片关联）==========
-- 小张收到张奶奶的各类消息（CHECK_IN 带照片 / NOT_REMIND 未打卡提醒 / BIND_SUCCESS 绑定成功）
INSERT INTO t_message (message_id,dedup_key,receiver_user_id,elder_user_id,type,params,ref_check_in_id,is_read,created_at,updated_at) VALUES
 ('MSG_1','checkin:CK_R1:U_GUARD_RICH','U_GUARD_RICH','U_ELDER_RICH','CHECK_IN','{"weather":"晴 26°C","city":"上海"}','CK_R1',0,@now,@now),
 ('MSG_2','checkin:CK_R2:U_GUARD_RICH','U_GUARD_RICH','U_ELDER_RICH','CHECK_IN','{"weather":"多云 24°C","city":"上海"}','CK_R2',1,@now-100000,@now),
 ('MSG_REMIND','remind:U_ELDER_RICH:U_GUARD_RICH:demo','U_GUARD_RICH','U_ELDER_RICH','NOT_REMIND','{"date":"今天","remindTime":"09:00"}','',0,@now-200000,@now),
 ('MSG_BIND','bind:G_RICH','U_GUARD_RICH','U_ELDER_RICH','BIND_SUCCESS','{"elderName":"张奶奶","relation":"GRANDMA"}','',1,@now-300000,@now);

-- ========== 9. 邀请（PENDING 待接受，测邀请流程）==========
SET @expire := @now + 7*24*3600*1000;
SET @past := @now - 7*24*3600*1000;
INSERT INTO t_invitation (invitation_id,invite_code,guardian_user_id,elder_phone,relation,remind_time,city,status,accepted_elder_user_id,accepted_at,expire_at,created_at,updated_at) VALUES
 -- PENDING 待接受（前端可用 elder_lonely 接受它，或 guardian_pending 撤销它）
 ('INV_PENDING','8888888888','U_GUARD_PENDING','13700137000','DAD','08:00','杭州','PENDING','',0,@expire,@now,@now),
 -- EXPIRED 已过期（测过期邀请展示）
 ('INV_EXPIRED','7777777777','U_GUARD_PENDING','13700137001','MOM','09:00','南京','EXPIRED','',0,@past,@past,@now),
 -- CANCELLED 已撤销
 ('INV_CANCELLED','6666666666','U_GUARD_PENDING','13700137002','GRANDPA','09:00','武汉','CANCELLED','',0,@expire,@now,@now),
 -- ACCEPTED 已接受（对应 guardian_rich → elder_rich 那条已成立的绑定）
 ('INV_ACCEPTED','5555555555','U_GUARD_RICH','13900139001','GRANDMA','09:00','上海','ACCEPTED','U_ELDER_RICH',@now,@expire,@now,@now);

SELECT '✅ 全场景数据已重建' AS result,
 (SELECT COUNT(*) FROM t_user) AS users,
 (SELECT COUNT(*) FROM t_fan) AS fans,
 (SELECT COUNT(*) FROM t_guardianship) AS guardianships,
 (SELECT COUNT(*) FROM t_elder_profile) AS profiles,
 (SELECT COUNT(*) FROM t_shipping_address) AS addresses,
 (SELECT COUNT(*) FROM t_check_in) AS checkins,
 (SELECT COUNT(*) FROM t_reward_task) AS tasks,
 (SELECT COUNT(*) FROM t_reward_claim) AS claims,
 (SELECT COUNT(*) FROM t_message) AS messages,
 (SELECT COUNT(*) FROM t_invitation) AS invitations;
