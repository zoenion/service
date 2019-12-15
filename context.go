package service

import (
	"context"
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	context2 "github.com/zoenion/service/context"
	"github.com/zoenion/service/discovery"
)

type contextKey string

const (
	ctxBox = contextKey("box")
)

func serviceBox(ctx context.Context) *Box {
	val := ctx.Value(ctxBox)
	if val == nil {
		srvCtx := ctx.Value(context2.ServiceContext)
		if srvCtx == nil {
			return nil
		}
		val = srvCtx.(context.Context).Value(ctxBox)
		if val == nil {
			return nil
		}
		return val.(*Box)
	}
	return val.(*Box)
}

func CACertificate(ctx context.Context) *x509.Certificate {
	box := serviceBox(ctx)
	if box == nil {
		return nil
	}
	return box.caCert
}

func Certificate(ctx context.Context) *x509.Certificate {
	box := serviceBox(ctx)
	if box == nil {
		return nil
	}
	return box.cert
}

func PrivateKey(ctx context.Context) crypto.PrivateKey {
	box := serviceBox(ctx)
	if box == nil {
		return nil
	}
	return box.privateKey
}

func Registry(ctx context.Context) discovery.Registry {
	box := serviceBox(ctx)
	if box == nil {
		return nil
	}
	return box.registry
}

func Namespace(ctx context.Context) string {
	box := serviceBox(ctx)
	if box == nil {
		return ""
	}
	return box.params.Namespace
}

func Name(ctx context.Context) string {
	box := serviceBox(ctx)
	if box == nil {
		return ""
	}
	return box.params.Name
}

func Dir(ctx context.Context) string {
	box := serviceBox(ctx)
	if box == nil {
		return ""
	}
	return box.params.Dir
}

func ID(ctx context.Context) string {
	box := serviceBox(ctx)
	if box == nil {
		return ""
	}
	return fmt.Sprintf("%s:%s", box.params.Namespace, box.params.Name)
}

func GatewayAddress(ctx context.Context, name string) string {
	box := serviceBox(ctx)
	if box == nil {
		return ""
	}
	gt, ok := box.gateways[name]
	if !ok {
		return ""
	}
	return gt.URL()
}

func ClientTLSConfig(ctx context.Context) *tls.Config {
	box := serviceBox(ctx)
	if box == nil {
		return nil
	}
	return box.ClientMutualTLS()
}

func RequestUser(ctx context.Context) (string, bool) {
	value := ctx.Value(context2.User)
	if user, ok := value.(string); ok {
		return user, true
	}
	return "", false
}

func ContextWithUser(ctx context.Context, user string) context.Context {
	return context.WithValue(ctx, context2.User, user)
}

func ContextWithBox(ctx context.Context, b *Box) context.Context {
	return context.WithValue(ctx, ctxBox, b)
}
