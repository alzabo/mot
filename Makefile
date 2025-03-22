GO_FILES := $(wildcard **/*.go)
YAML_Files := $(wildcard **/.y*ml)

test:
	addlicense -check -l apache -c "Ryan White" $(GO_FILES) $(YAML_FILES)

fix: addlicense

addlicense:
	addlicense -l apache -c "Ryan White" $(GO_FILES) $(YAML_FILES)

