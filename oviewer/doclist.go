package oviewer

import "fmt"

// DocumentLen returns the number of Docs.
func (root *Root) DocumentLen() int {
	root.mu.RLock()
	defer root.mu.RUnlock()
	return len(root.DocList)
}

// addDocument adds a document and displays it.
func (root *Root) addDocument(m *Document) {
	root.setMessagef("add %s", m.FileName)
	m.general = root.Config.General

	root.mu.Lock()
	root.DocList = append(root.DocList, m)
	root.CurrentDoc = len(root.DocList) - 1
	root.mu.Unlock()

	root.setDocument(m)
}

// closeDocument closes the document.
func (root *Root) closeDocument() {
	// If there is only one document, do nothing.
	if root.DocumentLen() == 1 {
		root.setMessage("only this document")
		return
	}

	root.setMessagef("close [%d]%s", root.CurrentDoc, root.Doc.FileName)

	root.mu.Lock()
	root.DocList = append(root.DocList[:root.CurrentDoc], root.DocList[root.CurrentDoc+1:]...)
	if root.CurrentDoc > 0 {
		root.CurrentDoc--
	}
	doc := root.DocList[root.CurrentDoc]
	root.mu.Unlock()

	root.setDocument(doc)
}

// nextDoc displays the next document.
func (root *Root) nextDoc() {
	root.setDocumentNum(root.CurrentDoc + 1)
	root.input.mode = Normal
	root.debugMessage("next document")
	root.debugMessage(fmt.Sprintf("cache %v\n", root.Doc.cache.Metrics.String()))
}

// previouseDoc displays the previous document.
func (root *Root) previousDoc() {
	root.setDocumentNum(root.CurrentDoc - 1)
	root.input.mode = Normal
	root.debugMessage("previous document")
	root.debugMessage(fmt.Sprintf("cache %v\n", root.Doc.cache.Metrics.String()))
}

// switchDocument displays the document of the specified docNum.
func (root *Root) switchDocument(docNum int) {
	root.setDocumentNum(docNum)
	root.debugMessage("switch document")
	root.debugMessage(fmt.Sprintf("cache %v\n", root.Doc.cache.Metrics.String()))
}

// setDocumentNum actually specifies docNum to display the document.
// This function is called internally from next / previous / switch / add.
func (root *Root) setDocumentNum(docNum int) {
	docNum = max(0, docNum)
	docNum = min(root.DocumentLen()-1, docNum)

	root.mu.Lock()
	root.CurrentDoc = docNum
	m := root.DocList[root.CurrentDoc]
	root.mu.Unlock()

	root.setDocument(m)
}

// setDocument sets the Document.
func (root *Root) setDocument(m *Document) {
	root.Doc = m
	root.ViewSync()
}

// helpDisplay is to switch between helpDisplay screen and normal screen.
func (root *Root) helpDisplay() {
	if root.screenMode == Help {
		root.toNormal()
		return
	}
	root.setDocument(root.helpDoc)
	root.screenMode = Help
}

// LogDisplay is to switch between Log screen and normal screen.
func (root *Root) logDisplay() {
	if root.screenMode == LogDoc {
		root.toNormal()
		return
	}
	root.setDocument(root.logDoc)
	root.screenMode = LogDoc
}

// toNormal displays a normal document.
func (root *Root) toNormal() {
	root.mu.RLock()
	m := root.DocList[root.CurrentDoc]
	root.mu.RUnlock()

	root.setDocument(m)
	root.screenMode = Docs
}
