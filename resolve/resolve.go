package resolve

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ipfs/go-namesys"
	path "github.com/ipfs/go-path"
	nsopt "github.com/ipfs/interface-go-ipfs-core/options/namesys"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// ErrNoNamesys is an explicit error for when an IPFS node doesn't
// (yet) have a name system
var ErrNoNamesys = errors.New(
	"core/resolve: no Namesys on IpfsNode - can't resolve ipns entry")

// ResolveIPNS resolves /ipns paths
func ResolveIPNS(ctx context.Context, nsys namesys.NameSystem, p path.Path, options ...nsopt.ResolveOpt) (path.Path, error) {
	ctx, span := namesys.StartSpan(ctx, "ResolveIPNS", trace.WithAttributes(attribute.String("Path", p.String())))
	defer span.End()
	if strings.HasPrefix(p.String(), "/ipns/") {
		// TODO(cryptix): we should be able to query the local cache for the path
		if nsys == nil {
			return "", ErrNoNamesys
		}

		seg := p.Segments()

		if len(seg) < 2 || seg[1] == "" { // just "/<protocol/>" without further segments
			err := fmt.Errorf("invalid path %q: ipns path missing IPNS ID", p)
			return "", err
		}

		extensions := seg[2:]
		resolvable, err := path.FromSegments("/", seg[0], seg[1])
		if err != nil {
			return "", err
		}

		respath, err := nsys.Resolve(ctx, resolvable.String(), options...)
		if err != nil {
			return "", err
		}

		segments := append(respath.Segments(), extensions...)
		p, err = path.FromSegments("/", segments...)
		if err != nil {
			return "", err
		}
	}
	return p, nil
}
