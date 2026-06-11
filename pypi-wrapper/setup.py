from setuptools import setup

VERSION = "1.0.10"

with open("README.md", "r", encoding="utf-8") as fh:
    long_description = fh.read()

setup(
    name="another-meet",
    version=VERSION,
    description="Manage Google Meet meetings from your terminal",
    long_description=long_description,
    long_description_content_type="text/markdown",
    author="parjanyaacoder",
    url="https://github.com/parjanyaacoder/another-meet",
    py_modules=["another_meet_wrapper"],
    entry_points={
        "console_scripts": [
            "another-meet=another_meet_wrapper:main",
        ],
    },
)
