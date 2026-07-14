// Package p_users implements the user administration and authentication system for the Lago framework.
// It manages users, roles, password hashes, session authentication, and route authorization.
//
// # Registrations and Features Added
//
// # Configurations
//
//   - "p_users.AuthConfig" -> p_users.AuthConfig
//     Defines system parameters for the initial administrator user (adminEmail, adminPassword).
//
// # Database Models
//
//   - p_users.User: DB model representing system users, including passwords, email, phone, and role references.
//   - p_users.Role: DB model representing access control roles (e.g. unassigned, superuser, admin).
//
// # Global Layers & Middlewares
//
//   - p_users.AuthenticationLayer: Validates "auth-token" session cookies and injects the active authenticated user into context.
//   - p_users.RoleAuthorizationLayer: Restricts access to downstream views depending on matching role membership.
//
// # Views
//
//   - "p_users.ListView": Renders list collection tables of all registered system users.
//   - "p_users.DetailView": Renders detail view of a target user.
//   - "p_users.CreateView": Processes creation and validation of new system user entries.
//   - "p_users.UpdateView": Processes updates on target system user records.
//   - "p_users.DeleteView": Processes deletion of target system user records.
//   - "p_users.ChangePasswordView": Superuser override handler for updating user passwords.
//   - "p_users.SelfDetailView": Renders details of the currently authenticated user profile.
//   - "p_users.SelfUpdateView": Processes profile updates for the currently logged-in user.
//   - "p_users.SelfChangePasswordView": Processes password changes for the currently logged-in user.
//   - "p_users.LoginView" & "p_users.SignupView" & "p_users.LogoutView": Handles session control, cookie allocation, and registration.
//   - "p_users.RoleListView" & "p_users.RoleDetailView" & "p_users.RoleCreateView" & "p_users.RoleUpdateView" & "p_users.RoleDeleteView": View actions targeting the user Roles table.
//
// # Routes
//
// Registers HTTP ServeMux path mappings:
//
//   - "/users/" -> p_users.ListView
//   - "/users/create/" -> p_users.CreateView
//   - "/users/u/{id}/" -> p_users.DetailView
//   - "/users/u/{id}/edit/" -> p_users.UpdateView
//   - "/users/u/{id}/delete/" -> p_users.DeleteView
//   - "/users/u/{id}/change-password/" -> p_users.ChangePasswordView
//   - "/users/self/" -> p_users.SelfDetailView
//   - "/users/self/edit/" -> p_users.SelfUpdateView
//   - "/users/self/change-password/" -> p_users.SelfChangePasswordView
//   - "/users/login/" -> p_users.LoginView
//   - "/users/signup/" -> p_users.SignupView
//   - "/users/logout/" -> p_users.LogoutView
//   - "/users/roles/" -> p_users.RoleListView
//   - "/users/roles/create/" -> p_users.RoleCreateView
//   - "/users/roles/{id}/" -> p_users.RoleDetailView
//   - "/users/roles/{id}/edit/" -> p_users.RoleUpdateView
//   - "/users/roles/{id}/delete/" -> p_users.RoleDeleteView
//
// # CLI Command Factories
//
//   - "p_users.createsuperuser": Cobra CLI command to manually generate a superuser account.
//   - "p_users.changepassword": Cobra CLI command to change user passwords by email.
//   - "p_users.revalidate_users": Cobra CLI command to sanitize and normalize users' email and phone formats.
//
// # Patches Applied
//
//   - "core.HomeRoute": Patches the default landing route "/" to render "core.HomeView" rather than a raw "Hello, World!" text.
package p_users
