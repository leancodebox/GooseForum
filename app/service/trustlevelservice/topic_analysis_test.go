package trustlevelservice

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/leancodebox/GooseForum/app/http/controllers/markdown2html"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
)

func TestTopicAnalysisReusesRenderedHTMLAndMetadata(t *testing.T) {
	content := "Description with **formatting**.\n\n[external](https://example.com) ![image](/topic.png)\n\n`![ignored](/code.png)`"
	for _, version := range []uint8{0, 1} {
		p := UserPolicy{Limits: pageConfig.TrustLimits{MaxImagesPerPost: new(1)}}
		got, err := p.AnalyzeTopic(content, version, true)
		if err != nil {
			t.Fatal(err)
		}
		metadata := markdown2html.AnalyzeContentVersion(content, 200, version)
		if got.RenderedHTML != markdown2html.PostMarkdownToHTMLVersion(content, version) || got.Description != metadata.Description || !reflect.DeepEqual(got.ImageURLs, metadata.ImageURLs) || got.FirstImageURL != metadata.FirstImageURL {
			t.Fatalf("analysis drift, version=%d: %+v", version, got)
		}
		p.Limits.MaxImagesPerPost = new(0)
		if _, err := p.AnalyzeTopic(content, version, true); err == nil {
			t.Fatal("image limit bypass")
		}
		if draft, err := p.AnalyzeTopic(content, version, false); err != nil || draft.RenderedHTML == "" {
			t.Fatalf("draft unexpectedly enforced publish limits: %+v err=%v", draft, err)
		}
		for _, published := range []bool{false, true} {
			p.Limits = pageConfig.TrustLimits{}
			fast, err := p.AnalyzeTopic(content, version, published)
			if err != nil || fast.RenderedHTML == "" || fast.Description != metadata.Description {
				t.Fatalf("unlimited fast path: %+v err=%v", fast, err)
			}
		}
	}
}

func BenchmarkTopicPolicyAnalysis(b *testing.B) {
	content := strings.Repeat("Paragraph with **bold**, `code` and a [link](https://example.com).\n\n", 200) + "![image](/topic.png)"
	p := UserPolicy{Limits: pageConfig.TrustLimits{MaxImagesPerPost: new(1)}}
	for _, version := range []uint8{0, 1} {
		b.Run(fmt.Sprintf("v%d/separate", version), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				if err := validateCounts(p.Limits, CountContent(content, version)); err != nil {
					b.Fatal(err)
				}
				analysis := markdown2html.AnalyzeContentVersion(content, 200, version)
				if analysis.Description == "" {
					b.Fatal("empty description")
				}
			}
		})
		b.Run(fmt.Sprintf("v%d/shared", version), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				analysis, err := p.AnalyzeTopic(content, version, true)
				if err != nil || analysis.RenderedHTML == "" || analysis.Description == "" {
					b.Fatal("analysis failed", err)
				}
			}
		})
	}
}
