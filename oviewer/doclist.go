package oviewer

import (
	"context"
	"slices"
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

// hasDocChanged checks if any document in the list has changed.
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
func (root *Root) addDocument(ctx context.Context, addDoc *Document) {
	root.mu.Lock()
	root.DocList = append(root.DocList, addDoc)
	root.mu.Unlock()

	root.setDocumentNum(ctx, len(root.DocList)-1)
	root.setMessageLogf("add %s%s", addDoc.FileName, addDoc.Caption)
}

// insertDocument inserts a document after the specified number and displays it.
func (root *Root) insertDocument(ctx context.Context, num int, m *Document) {
	root.mu.Lock()
	num = max(0, num)
	num = min(len(root.DocList)-1, num)
	root.DocList = append(root.DocList[:num+1], append([]*Document{m}, root.DocList[num+1:]...)...)
	root.mu.Unlock()

	go root.waitForEOF(m)

	root.setDocumentNum(ctx, num+1)
	root.setMessageLogf("insert %s%s", m.FileName, m.Caption)
}

// closeDocument closes the document.
func (root *Root) closeDocument(ctx context.Context) {
	// If there is only one document, do nothing.
	if root.DocumentLen() == 1 {
		root.setMessage("only this document")
		return
	}

	root.setMessageLogf("close [%d]%s%s", root.CurrentDoc, root.Doc.FileName, root.Doc.Caption)

	root.mu.Lock()
	num := root.CurrentDoc
	root.DocList[num].requestClose()
	root.DocList = slices.Delete(root.DocList, num, num+1)
	if num > 0 {
		num--
	}
	root.mu.Unlock()

	root.setDocumentNum(ctx, num)
}

// closeAllDocumentsOfType closes all documents of the specified type.
func (root *Root) closeAllDocumentsOfType(dType documentType) (int, []string) {
	root.mu.Lock()
	defer root.mu.Unlock()

	docLen := len(root.DocList)
	docNum := root.CurrentDoc
	closed := make([]string, 0, docLen)
	for i := docLen - 1; i >= 0; i-- {
		if len(root.DocList) <= 1 {
			break
		}
		doc := root.DocList[i]
		if doc.documentType == dType {
			doc.requestClose()
			root.DocList = slices.Delete(root.DocList, i, i+1)
			closed = append(closed, doc.FileName+doc.Caption)
			if docNum == i {
				docNum--
			}
		}
	}
	return docNum, closed
}

// nextDoc displays the next document.
func (root *Root) nextDoc(ctx context.Context) {
	root.setDocumentNum(ctx, root.CurrentDoc+1)
	root.input.Event = normal()
	root.debugMessage("next document")
}

// previousDoc displays the previous document.
func (root *Root) previousDoc(ctx context.Context) {
	fromNum := root.CurrentDoc
	fromDoc := root.getDocument(fromNum)
	if fromDoc == nil {
		return
	}
	toNum := root.CurrentDoc - 1
	toDoc := root.getDocument(toNum)
	if toDoc == nil {
		return
	}

	root.setDocumentNum(ctx, toNum)
	root.input.Event = normal()
	root.debugMessage("previous document")

	if fromDoc.parent != toDoc {
		return
	}
	root.linkLineNum(fromDoc, toDoc)
}

// linkLineNum links the line number of the parent document.
func (root *Root) linkLineNum(fromDoc, toDoc *Document) {
	if fromDoc.lineNumMap == nil {
		return
	}
	if n, ok := fromDoc.lineNumMap.LoadForward(fromDoc.topLN + fromDoc.firstLine()); ok {
		root.debugMessage("Move parent line number")
		toDoc.moveLine(n - toDoc.firstLine())
	}
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
	if root.CurrentDoc < 0 || root.CurrentDoc >= len(root.DocList) {
		return
	}
	m := root.DocList[root.CurrentDoc]
	root.setDocument(ctx, m)
}
