// Package audit @Author larry
// @Date 2025/11/13 14:40
// @Desc

package impls

import (
	"github.com/sirupsen/logrus"

	"warm-nest/pkg/tool/machines/impls/audits"
	"warm-nest/pkg/tool/machines/model/context"
	"warm-nest/pkg/tool/machines/model/manual"
)

type IMachineAudit interface {
	Audit(ctx *context.MachineContext, cfg manual.ManualConfig, event manual.ManualEvent, authFunc manual.AuthFunc) error
}

type MachineAudit struct {
	authFuncs []manual.AuthFunc
}

func NewAudit(authFuncs ...manual.AuthFunc) MachineAudit {
	machineAudit := MachineAudit{
		authFuncs: []manual.AuthFunc{
			audits.AuditAction, audits.AuditUser, audits.AuditRole, audits.AuditExpr,
		},
	}
	machineAudit.authFuncs = append(machineAudit.authFuncs, authFuncs...)
	return machineAudit
}

func (e *MachineAudit) Audit(ctx *context.MachineContext, cfg manual.ManualConfig, event manual.ManualEvent) error {
	if !cfg.DisableSupper && event.Operator.Supper {
		ctx.Log().WithFields(logrus.Fields{
			"event": event,
		}).Info("operator is supper! skip audit")
		return nil
	}

	authFuncs := append(e.authFuncs, cfg.GetAuthFunc(event.Action)...)
	for index, af := range authFuncs {
		if err := af(ctx, cfg, event); err != nil {
			ctx.Log().WithFields(logrus.Fields{
				"index": index,
				"cfg":   cfg,
				"event": event,
			}).WithError(err).Error("audit not pass")
			return err
		}
	}
	return nil
}
