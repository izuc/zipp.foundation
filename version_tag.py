import os
import subprocess

# Get the directory of the currently executing script
script_dir = os.path.dirname(os.path.abspath(__file__))

# Function to tag a module
def tag_module(module_name, version):
    tag_name = "v" + version  # You can customize the tag format
    subprocess.run(["git", "tag", "-f", tag_name])
    subprocess.run(["git", "push", "origin", tag_name])  # Push the tag to the remote repository

# Initialize a counter for versioning
version_counter = 1

# Iterate through directories in the script's directory
for dirpath, dirnames, filenames in os.walk(script_dir):
    if "go.mod" in filenames:
        # Extract the module name from the go.mod file
        module_name = ""
        with open(os.path.join(dirpath, "go.mod"), "r") as go_mod_file:
            for line in go_mod_file:
                if line.startswith("module "):
                    module_name = line.strip().split(" ")[1]
                    break  # Exit loop once module name is found

        # If module_name is empty, skip the rest of the loop
        if not module_name:
            continue

        # Set the base version
        base_version = "0.1.2"

        # Tag the module
        print(f"Tagging {module_name} with version {base_version}")
        tag_module(module_name, base_version)
