# Basic user with minimal required fields
resource "gitea_user" "basic" {
  username = "basicuser"
  email    = "basic@example.com"
  password = "SecurePassword123!"
}

# User with full name and profile information
resource "gitea_user" "profile_complete" {
  username  = "johndoe"
  email     = "john.doe@example.com"
  password  = "SecurePassword123!"
  full_name = "John Doe"
}

# User with restricted access
resource "gitea_user" "restricted_user" {
  username   = "restricteduser"
  email      = "restricted@example.com"
  password   = "SecurePassword123!"
  full_name  = "Restricted User"
  restricted = true
  visibility = "limited"
}

# User with custom visibility (private)
resource "gitea_user" "private_user" {
  username   = "privateuser"
  email      = "private@example.com"
  password   = "SecurePassword123!"
  full_name  = "Private User"
  visibility = "private"
}

# User with custom visibility (public)
resource "gitea_user" "public_user" {
  username   = "publicuser"
  email      = "public@example.com"
  password   = "SecurePassword123!"
  full_name  = "Public User"
  visibility = "public"
}

# User that must change password on first login
resource "gitea_user" "must_change_pwd" {
  username             = "tempuser"
  email                = "temp@example.com"
  password             = "TemporaryPass123!"
  full_name            = "Temporary User"
  must_change_password = true
}

# User with welcome notification disabled
resource "gitea_user" "no_notification" {
  username    = "silentuser"
  email       = "silent@example.com"
  password    = "SecurePassword123!"
  full_name   = "Silent User"
  send_notify = false
}

# User with custom authentication source
# resource "gitea_user" "external_auth" {
#   username  = "externaluser"
#   email     = "external@example.com"
#   password  = "SecurePassword123!"
#   full_name = "External Auth User"
#   source_id = 1
# }

# User with explicit creation timestamp (migration scenario)
resource "gitea_user" "migrated_user" {
  username   = "migrateduser"
  email      = "migrated@example.com"
  password   = "SecurePassword123!"
  full_name  = "Migrated User"
  created_at = "2020-01-01T00:00:00Z"
}

# User with all optional fields set
resource "gitea_user" "complete_profile" {
  username             = "poweruser"
  email                = "power@example.com"
  password             = "SecurePassword123!"
  full_name            = "Power User"
  visibility           = "public"
  restricted           = false
  must_change_password = false
  send_notify          = true
}

# User with limited visibility and restricted access
resource "gitea_user" "limited_restricted" {
  username   = "limiteduser"
  email      = "limited@example.com"
  password   = "SecurePassword123!"
  full_name  = "Limited User"
  visibility = "limited"
  restricted = true
}

# User for testing account with temporary password
resource "gitea_user" "test_account" {
  username             = "testaccount"
  email                = "test@example.com"
  password             = "TestPassword123!"
  full_name            = "Test Account"
  must_change_password = true
  send_notify          = false
}

# Bot/Service account user
resource "gitea_user" "service_bot" {
  username    = "servicebot"
  email       = "bot@example.com"
  password    = "BotPassword123!"
  full_name   = "Service Bot"
  visibility  = "private"
  send_notify = false
}
