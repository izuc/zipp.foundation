import os

def process_file(file_path):
    with open(file_path, 'r') as file:
        lines = file.readlines()

    updated_lines = []

    for line in lines:
        stripped_line = line.strip()
        if "github.com/izuc/zipp.foundation/" in stripped_line and not stripped_line.startswith("module "):
            parts = stripped_line.split()
            module_name = parts[0]
            # Ensure the line contains the module name and not some other line that matches
            if module_name.startswith("github.com/izuc/zipp.foundation/"):
                updated_line = "\t" + module_name + " v0.1.2\n"
            else:
                updated_line = line
        else:
            updated_line = line
        updated_lines.append(updated_line)

    # Write the updated lines back to the file
    with open(file_path, 'w') as file:
        file.writelines(updated_lines)


def main():
    for subdir, _, files in os.walk('.'):
        for file in files:
            if file == 'go.mod':
                process_file(os.path.join(subdir, file))
                print(f"Processed: {os.path.join(subdir, file)}")


if __name__ == "__main__":
    main()
