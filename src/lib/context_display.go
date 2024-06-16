package lib

import "desktop-cleaner/internal"

func GetContextLabelAndIcon(contextType internal.ContextType) (string, string) {
	var icon string
	var lbl string
	switch contextType {
	case internal.ContextFileType:
		icon = "📄"
		lbl = "file"
	case internal.ContextURLType:
		icon = "🌎"
		lbl = "url"
	case internal.ContextDirectoryTreeType:
		icon = "🗂 "
		lbl = "tree"
	case internal.ContextNoteType:
		icon = "✏️ "
		lbl = "note"
	case internal.ContextPipedDataType:
		icon = "↔️ "
		lbl = "piped"
	case internal.ContextImageType:
		icon = "🖼️ "
		lbl = "image"
	}

	return lbl, icon
}
