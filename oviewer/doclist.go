package oviewer

import (
	"context"
	"sync/atomic"
)

// DocumentLen returns the number of Docs.
func (root *Root) DocumentLen() int {
	root.mu.RLock()
	defer root.mu.RUnlock()
	return len(root.DocList)
}

// getDocument returns the document of the specified docNum.
func (root *Root) getDocument(docNum int) *Document {
	root.mu.RLock()
	defer root.mu.RUnlock()
	if docNum < 0 || docNum >= len(root.DocList) {
		return nil
	}
	return root.DocList[docNum]
}

// hasDocChanged() returns if doc has changed.
func (root *Root) hasDocChanged() bool {
	root.mu.RLock()
	defer root.mu.RUnlock()
	eventFlag := false
	for _, doc := range root.DocList {
		if atomic.SwapInt32(&doc.store.changed, 0) == 1 {
			eventFlag = true
		}
	}
	return eventFlag
}

// addDocument adds a document and displays it.
func (root *Root) addDocument(ctx context.Context, m *Document) {
	root.setMessageLogf("add %s", m.FileName)
	m.general = root.Config.General
	m.regexpCompile()

	root.mu.Lock()
	defer root.mu.Unlock()
	root.DocList = append(root.DocList, m)
	root.CurrentDoc = len(root.DocList) - 1
	root.showDocNum = true

	root.setDocument(ctx, m)
}

// closeDocument closes the document.
func (root *Root) closeDocument(ctx context.Context) {
	// If there is only one document, do nothing.
	if root.DocumentLen() == 1 {
		root.setMessage("only this document")
		return
	}

	root.setMessageLogf("close [%d]%s", root.CurrentDoc, root.Doc.FileName)
	root.mu.Lock()
	defer root.mu.Unlock()
	root.DocList[root.CurrentDoc].requestClose()
	root.DocList = append(root.DocList[:root.CurrentDoc], root.DocList[root.CurrentDoc+1:]...)
	if root.CurrentDoc > 0 {
		root.CurrentDoc--
	}
	doc := root.DocList[root.CurrentDoc]
	root.setDocument(ctx, doc)
}

// closeAllDocument closes all documents of the specified type.
func (root *Root) closeAllDocument(ctx context.Context, dType documentType) {
	root.mu.Lock()
	for i := len(root.DocList) - 1; i >= 0; i-- {
		if len(root.DocList) <= 1 {
			break
		}
		doc := root.DocList[i]
		if doc.documentType == dType {
			root.DocList = append(root.DocList[:i], root.DocList[i+1:]...)
			root.setMessageLogf("close %s", doc.FileName)
		}
	}
	if root.CurrentDoc >= len(root.DocList) {
		root.CurrentDoc = len(root.DocList) - 1
	}
	doc := root.DocList[root.CurrentDoc]
	root.mu.Unlock()
	root.setDocument(ctx, doc)
}

// nextDoc displays the next document.
func (root *Root) nextDoc(ctx context.Context) {
	root.setDocumentNum(ctx, root.CurrentDoc+1)
	root.input.Event = normal()
	root.debugMessage("next document")
}

// previousDoc displays the previous document.
func (root *Root) previousDoc(ctx context.Context) {
	lineNum := 0
	targetNum := root.CurrentDoc - 1
	targetDoc := root.getDocument(targetNum)
	if targetDoc == root.Doc.parent && root.Doc.lineNumMap != nil {
		if n, ok := root.Doc.lineNumMap.LoadForward(root.Doc.topLN + root.Doc.firstLine()); ok {
			lineNum = n
		}
	}
	root.setDocumentNum(ctx, root.CurrentDoc-1)
	root.input.Event = normal()
	if lineNum > 0 {
		root.sendGoto(lineNum - root.Doc.firstLine() + 1)
	}
	root.debugMessage("previous document")
}

// switchDocument displays the document of the specified docNum.
func (root *Root) switchDocument(ctx context.Context, docNum int) {
	root.setDocumentNum(ctx, docNum)
	root.debugMessage("switch document")
}

// setDocumentNum actually specifies docNum to display the document.
// This function is called internally from next / previous / switch / add.
func (root *Root) setDocumentNum(ctx context.Context, docNum int) {
	docNum = max(0, docNum)
	docNum = min(root.DocumentLen()-1, docNum)

	root.mu.Lock()
	defer root.mu.Unlock()

	root.CurrentDoc = docNum
	m := root.DocList[root.CurrentDoc]
	root.setDocument(ctx, m)
}

// setDocument sets the Document.
func (root *Root) setDocument(ctx context.Context, m *Document) {
	root.Doc = m
	root.ViewSync(ctx)
}

// helpDisplay is to switch between helpDisplay screen and normal screen.
func (root *Root) helpDisplay(ctx context.Context) {
	if root.Doc.documentType == DocHelp {
		root.toNormal(ctx)
		return
	}
	root.setDocument(ctx, root.helpDoc)
}

// LogDisplay is to switch between Log screen and normal screen.
func (root *Root) logDisplay(ctx context.Context) {
	if root.Doc.documentType == DocLog {
		root.toNormal(ctx)
		return
	}
	root.setDocument(ctx, root.logDoc.Document)
}

// toNormal displays a normal document.
func (root *Root) toNormal(ctx context.Context) {
	root.mu.RLock()
	defer root.mu.RUnlock()

	m := root.DocList[root.CurrentDoc]
	root.setDocument(ctx, m)
}
