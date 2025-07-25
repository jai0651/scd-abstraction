import sys
from jinja2 import Environment, FileSystemLoader
import os

def generate(target):
    env = Environment(loader=FileSystemLoader('.'))
    template = env.get_template('scd_helper_template.jinja')
    rendered = template.render(target=target)
    os.makedirs('scd', exist_ok=True)
    if target == 'django':
        output_file = os.path.join('python', 'django_example', 'scd', 'scd_helpers.py')
    elif target == 'gorm':
        output_file = os.path.join('Go', 'scd', 'scd_helpers.go')
    else:
        print("Unknown target. Supported: django, gorm")
        sys.exit(1)
    with open(output_file, 'w') as f:
        f.write(rendered)
    print(f"Generated {output_file} for {target}")

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage: python generate_scd_helper.py <target>")
        print("Targets: django, gorm")
        sys.exit(1)
    target = sys.argv[1]
    generate(target) 