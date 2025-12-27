# Output file
Output demo.gif

# Settings
Set FontSize 16
Set Width 1600
Set Height 800
Set Padding 20
Set FontFamily "Hack Nerd Font"
Set Theme "Argonaut"

# Start the application
Type "fresh --dir ~/git"
Sleep 1s
Enter

# Allow some time for scanning/loading
Sleep 3s

# Simulate navigation
Type "j"
Sleep 800ms
Type "j"
Sleep 800ms
Type "j"
Sleep 800ms
Type "k"
Sleep 800ms
Type "k"
Sleep 4s

# Simulate Pull All
Ctrl+P
Sleep 7s

# Quit the application
Type "q"
Sleep 1s
