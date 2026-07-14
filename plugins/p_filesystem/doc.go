// Package p_filesystem implements a virtual node (VNode) database filesystem for the Lago framework.
// It integrates abstract storage backends (local directory storage or Google Cloud Storage) with database VNode entities
// to handle file uploads, downloads, folders hierarchies, and image displays.
//
// # Registrations and Features Added
//
// # Configurations
//
//   - "p_filesystem.FilesystemConfig" -> p_filesystem.FilesystemConfig
//     Configures storage backends (local directories or Google Cloud Storage buckets, credentials, prefixes).
//
// # Database Models
//
//   - p_filesystem.VNode: DB model mapping file/directory nodes, parent links, file sizes, and MIME upload files.
//
// # Pages
//
//   - p_filesystem.VNodeTable & p_filesystem.VNodeDetail: Renders lists, directories, and files details.
//   - p_filesystem.VNodeCreateForm & p_filesystem.VNodeUpdateForm: Form controls for directory creations and file uploads.
//   - p_filesystem.VNodeSelectionTable: Selection lists targeting files.
//
// # Routes
//
// Registers HTTP ServeMux path mappings:
//
//   - "/filesystem/" -> p_filesystem.ListView
//   - "/filesystem/create/" -> p_filesystem.CreateView
//   - "/filesystem/u/{id}/" -> p_filesystem.DetailView
//   - "/filesystem/u/{id}/edit/" -> p_filesystem.UpdateView
//   - "/filesystem/u/{id}/delete/" -> p_filesystem.DeleteView
//   - "/filesystem/select/" -> p_filesystem.SelectView
//
// # Views
//
//   - "p_filesystem.ListView" & "p_filesystem.DetailView": Renders file/directory hierarchies.
//   - "p_filesystem.CreateView" & "p_filesystem.UpdateView" & "p_filesystem.DeleteView": Processes uploads, directory generation, and file deletion.
//   - "p_filesystem.SelectView": Renders list collection file selector components.
//
// # Seeding Generators
//
//   - "p_filesystem.generators": Registers random folder/file seeder generators for building mock filesystem records under testing.
package p_filesystem
