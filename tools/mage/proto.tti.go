// Copyright © 2019 The Things Industries B.V.

package ttnmage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/magefile/mage/target"
	"golang.org/x/xerrors"
)

// TTIProto namespace.
type TTIProto mg.Namespace

// Go generates Go protos.
func (p TTIProto) Go(context.Context) error {
	if err := withProtoc(func(pCtx *protocContext, protoc func(...string) error) error {
		var convs []string
		for _, t := range []string{"any", "duration", "empty", "field_mask", "struct", "timestamp", "wrappers"} {
			convs = append(convs, fmt.Sprintf("Mgoogle/protobuf/%s.proto=github.com/gogo/protobuf/types", t))
		}
		convStr := strings.Join(convs, ",")

		if err := protoc(
			fmt.Sprintf("--fieldmask_out=lang=gogo,%s:%s", convStr, protocOut),
			fmt.Sprintf("--gogottn_out=plugins=grpc,%s:%s", convStr, protocOut),
			fmt.Sprintf("--grpc-gateway_out=%s:%s", convStr, protocOut),
			fmt.Sprintf("%s/api/tti/*.proto", pCtx.WorkingDirectory),
		); err != nil {
			return xerrors.Errorf("failed to generate protos: %w", err)
		}
		return nil
	}); err != nil {
		return err
	}

	if err := sh.RunV(filepath.Join("tools", "mage", "scripts", "fix-grpc-gateway-names.sh"), "api", "api/tti"); err != nil {
		return xerrors.Errorf("failed to fix gRPC-gateway names: %w", err)
	}

	ttipb, err := filepath.Abs(filepath.Join("pkg", "ttipb"))
	if err != nil {
		return xerrors.Errorf("failed to construct absolute path to pkg/ttipb: %w", err)
	}
	if err := runGoTool("golang.org/x/tools/cmd/goimports", "-w", ttipb); err != nil {
		return xerrors.Errorf("failed to run goimports on generated code: %w", err)
	}
	if err := runUnconvert(ttipb); err != nil {
		return xerrors.Errorf("failed to run unconvert on generated code: %w", err)
	}
	return sh.RunV("gofmt", "-w", "-s", ttipb)
}

// GoClean removes generated Go protos.
func (p TTIProto) GoClean(context.Context) error {
	return filepath.Walk(filepath.Join("pkg", "ttipb"), func(path string, _ os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		for _, ext := range []string{".pb.go", ".pb.gw.go", ".pb.fm.go", ".pb.util.go"} {
			if strings.HasSuffix(path, ext) {
				if err := sh.Rm(path); err != nil {
					return err
				}
				return nil
			}
		}
		return nil
	})
}

// Swagger generates Swagger protos.
func (p TTIProto) Swagger(context.Context) error {
	changed, err := target.Glob(filepath.Join("api", "tti", "api.swagger.json"), filepath.Join("api", "tti", "*.proto"))
	if err != nil {
		return xerrors.Errorf("failed checking modtime: %w", err)
	}
	if !changed {
		return nil
	}
	return withProtoc(func(pCtx *protocContext, protoc func(...string) error) error {
		if err := protoc(
			fmt.Sprintf("--swagger_out=allow_merge,merge_file_name=api:%s/api/tti", pCtx.WorkingDirectory),
			fmt.Sprintf("%s/api/tti/*.proto", pCtx.WorkingDirectory),
		); err != nil {
			return xerrors.Errorf("failed to generate protos: %w", err)
		}
		return nil
	})
}

// SwaggerClean removes generated Swagger protos.
func (p TTIProto) SwaggerClean(context.Context) error {
	return sh.Rm(filepath.Join("api", "tti", "api.swagger.json"))
}

// Markdown generates Markdown protos.
func (p TTIProto) Markdown(context.Context) error {
	changed, err := target.Glob(filepath.Join("api", "tti", "api.md"), filepath.Join("api", "tti", "*.proto"))
	if err != nil {
		return xerrors.Errorf("failed checking modtime: %w", err)
	}
	if !changed {
		return nil
	}
	return withProtoc(func(pCtx *protocContext, protoc func(...string) error) error {
		if err := protoc(
			fmt.Sprintf("--doc_opt=%s/api/api.md.tmpl,api.md --doc_out=%s/api/tti", pCtx.WorkingDirectory, pCtx.WorkingDirectory),
			fmt.Sprintf("%s/api/tti/*.proto", pCtx.WorkingDirectory),
		); err != nil {
			return xerrors.Errorf("failed to generate protos: %w", err)
		}
		return nil
	})
}

// MarkdownClean removes generated Markdown protos.
func (p TTIProto) MarkdownClean(context.Context) error {
	return sh.Rm(filepath.Join("api", "tti", "api.md"))
}

// JsSDK generates javascript SDK protos.
func (p TTIProto) JsSDK(context.Context) error {
	changed, err := target.Glob(filepath.Join("sdk", "js", "generated", "api.tti.json"), filepath.Join("api", "tti", "*.proto"))
	if err != nil {
		return xerrors.Errorf("failed checking modtime: %w", err)
	}
	if !changed {
		return nil
	}
	return withProtoc(func(pCtx *protocContext, protoc func(...string) error) error {
		if err := protoc(
			fmt.Sprintf("--doc_opt=json,api.tti.json --doc_out=%s/v3/sdk/js/generated", pCtx.WorkingDirectory),
			fmt.Sprintf("%s/api/tti/*.proto", pCtx.WorkingDirectory),
		); err != nil {
			return xerrors.Errorf("failed to generate protos: %w", err)
		}
		return nil
	})
}

// JsSDKClean removes generated javascript SDK protos.
func (p TTIProto) JsSDKClean(context.Context) error {
	return sh.Rm(filepath.Join("sdk", "js", "generated", "api.tti.json"))
}

// All generates protos.
func (p TTIProto) All(ctx context.Context) {
	mg.CtxDeps(ctx, TTIProto.Go, TTIProto.Swagger, TTIProto.Markdown, TTIProto.JsSDK)
}

// Clean removes generated protos.
func (p TTIProto) Clean(ctx context.Context) {
	mg.CtxDeps(ctx, TTIProto.GoClean, TTIProto.SwaggerClean, TTIProto.MarkdownClean, TTIProto.JsSDKClean)
}
