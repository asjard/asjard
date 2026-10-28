package main

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/compiler/protogen"
)

func TestGetCommentTitleAndDesc(t *testing.T) {
	g := &McpGenerator{}
	t.Run("empty", func(t *testing.T) {
		comments := protogen.CommentSet{}
		require.Empty(t, g.getCommentTitle(comments))
		require.Empty(t, g.getCommentDesc(comments))
	})
	t.Run("one line", func(t *testing.T) {
		comments := protogen.CommentSet{Leading: " SayHello returns a greeting.\n"}
		require.Equal(t, "SayHello returns a greeting.", g.getCommentTitle(comments))
		require.Equal(t, "SayHello returns a greeting.", g.getCommentDesc(comments))
	})
	t.Run("multiple lines", func(t *testing.T) {
		comments := protogen.CommentSet{
			Leading: " SayHello returns a greeting.\n It supports multiple HTTP GET entrypoints.\n",
		}
		require.Equal(t, "SayHello returns a greeting.", g.getCommentTitle(comments))
		require.Equal(t, "It supports multiple HTTP GET entrypoints.", g.getCommentDesc(comments))
	})
}

func TestJoinCommentAndRules(t *testing.T) {
	require.Equal(t, "源人脸图片的公开可访问 URL [required]", joinCommentAndRules("源人脸图片的公开可访问 URL", "required"))
	require.Equal(t, "源人脸图片的公开可访问 URL", joinCommentAndRules("源人脸图片的公开可访问 URL", ""))
	require.Equal(t, "required,max=20", joinCommentAndRules("", "required,max=20"))
}

func TestRuleHasRequired(t *testing.T) {
	require.True(t, ruleHasRequired("required,max=20"))
	require.True(t, ruleHasRequired("max=10;Create:required"))
	require.False(t, ruleHasRequired("max=20"))
	require.False(t, ruleHasRequired("required_if=Foo"))
}
