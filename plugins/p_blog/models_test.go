package p_blog

import (
	"testing"

	"github.com/lariv-in/lariv/plugins/p_users"
)

func TestBlogAndBlogTagModels(t *testing.T) {
	tag1 := BlogTag{
		Name: "tech.golang",
	}

	tag2 := BlogTag{
		Name: "tech.web",
	}

	user := p_users.User{
		Name:  "Author User",
		Email: "author@example.com",
	}
	user.ID = 42

	blog := Blog{
		Title:       "Introduction to Lariv Plugins",
		Description: "A comprehensive guide on building plugins for Lariv.",
		CreatedByID: user.ID,
		CreatedBy:   user,
		Content:     "# Welcome\nThis is a markdown content article.",
		Tags:        []BlogTag{tag1, tag2},
	}

	if err := blog.BeforeSave(nil); err != nil {
		t.Fatalf("unexpected error in BeforeSave: %v", err)
	}

	if blog.Title != "Introduction to Lariv Plugins" {
		t.Errorf("expected Title 'Introduction to Lariv Plugins', got %q", blog.Title)
	}

	if blog.Slug != "introduction-to-lariv-plugins" {
		t.Errorf("expected auto-generated Slug 'introduction-to-lariv-plugins', got %q", blog.Slug)
	}

	if blog.Description != "A comprehensive guide on building plugins for Lariv." {
		t.Errorf("expected Description 'A comprehensive guide on building plugins for Lariv.', got %q", blog.Description)
	}

	if blog.CreatedByID != 42 {
		t.Errorf("expected CreatedByID 42, got %d", blog.CreatedByID)
	}

	if blog.CreatedBy.Email != "author@example.com" {
		t.Errorf("expected CreatedBy Email 'author@example.com', got %q", blog.CreatedBy.Email)
	}

	if blog.Content != "# Welcome\nThis is a markdown content article." {
		t.Errorf("expected markdown Content, got %q", blog.Content)
	}

	if len(blog.Tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(blog.Tags))
	}

	if blog.Tags[0].Name != "tech.golang" {
		t.Errorf("expected tag[0] Name 'tech.golang', got %q", blog.Tags[0].Name)
	}

	if blog.Tags[1].Name != "tech.web" {
		t.Errorf("expected tag[1] Name 'tech.web', got %q", blog.Tags[1].Name)
	}
}
