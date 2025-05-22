package example

# By default deny
default allow := false

# Allow if the user is "alice"
allow if {
  input.user == "alice"
}
