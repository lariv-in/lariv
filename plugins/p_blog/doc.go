// Package p_blog provides blog post management and hierarchical tagging capabilities for Lariv.
//
// # Models
//
//   - Blog: Represents a blog article containing a Title, a foreign key CreatedBy pointing to p_users.User, Markdown Content, and many-to-many Tags relationship.
//   - BlogTag: Represents hierarchical blog tags using PostgreSQL ltree datatype and many-to-many Blogs relationship.
package p_blog
