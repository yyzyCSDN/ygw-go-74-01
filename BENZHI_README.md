基于 Go 实现的洁净室/手术室环境联控系统项目，一款医疗环境保障服务，完成压差监测与正压维持、粒子计数采样、送排风联动、门禁联锁与消毒流程管理。

# CleanroomORControl

CleanroomORControl 是自包含的洁净室/手术室环境联控服务。系统持续采集房间
压差、粒子计数与温湿度数据，通过送排风机组联动维持正压与洁净度，管理消毒
流程的门禁联锁与通风复位，并在异常时向护士站告警。

## 构建

```bash
go build -mod=vendor ./...
```

## 运行

```bash
go run -mod=vendor ./cmd/cleanroom -addr 127.0.0.1:8090 -dir ./data
```

启动后访问 http://127.0.0.1:8090/ 打开控制台页面。

## HTTP 接口

- `GET /healthz` 健康检查
- `GET /api/v1/status` 房间、压差、送排风、门禁与消毒状态总览
- `POST /api/v1/pressure/sample` 提交压差采样
- `POST /api/v1/particle/sample` 提交粒子计数采样
- `POST /api/v1/disinfection/start` 开始消毒
- `POST /api/v1/disinfection/complete` 标记消毒阶段完成
- `POST /api/v1/ventilation` 执行消毒后通风
- `POST /api/v1/fan/switch` 切换送风或排风机组
- `POST /api/v1/env/sample` 提交温湿度采样
- `GET /api/v1/alarms` 查询告警记录
- `GET /api/v1/doors` 查询门禁联锁状态
- `POST /api/v1/persist` 持久化压差快照
- `POST /api/v1/settle` 模拟气流稳定信号

## 数据目录

运行目录下的 `data/records/` 保存粒子采样记录，`data/meta/` 保存重启恢复
用的压差快照。服务重启时会读取最新快照恢复告警状态。
