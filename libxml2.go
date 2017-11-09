package xsdvalidate

/*
#cgo pkg-config: libxml-2.0
#include <string.h>
#include <libxml/xmlschemastypes.h>
#include <errno.h>
#include <malloc.h>
#include <stdbool.h>
#define GO_ERR_INIT 512
#define P_ERR_DEFAULT 1
#define P_ERR_EXT 2
#define LIBXML_STATIC

struct xsdParserResult {
	xmlSchemaPtr schemaPtr;
	char *errorStr;
};

struct xmlParserResult {
	xmlDocPtr docPtr;
	char *errorStr;
};

struct errCtx {
	char *errBuf;
};

struct simpleXmlError {
	int	code;
	char*	message;
	int 	level;
	int	line;
	char*	node;
};

static void noOutputCallback(void *ctx, const char *message, ...) {}

static void init() {
	xmlInitParser();
}

static void cleanup() {
	xmlSchemaCleanupTypes();
	xmlCleanupParser();
}

static void genErrorCallback(void *ctx, const char *message, ...) {
	struct errCtx *ectx = ctx;
	char *newLine = malloc(GO_ERR_INIT);

	va_list varArgs;
        va_start(varArgs, message);

	int oldLen = strlen(ectx->errBuf) + 1;
	int lineLen = 1 + vsnprintf(newLine, GO_ERR_INIT, message, varArgs);

	if (lineLen  > GO_ERR_INIT) {
		va_end(varArgs);
		va_start(varArgs, message);
		free(newLine);
		newLine = malloc(lineLen);
		vsnprintf(newLine, lineLen, message, varArgs);
		va_end(varArgs);
	} else {
		va_end(varArgs);
	}

	char *tmp = malloc(oldLen + lineLen);
	memcpy(tmp, ectx->errBuf, oldLen);
	strcat(tmp, newLine);
	free(newLine);
	free(ectx->errBuf);
	ectx->errBuf = tmp;
}

static void simpleStructErrorCallback(void *ctx, xmlErrorPtr p) {
	struct simpleXmlError *sErr = ctx;
	sErr->code = p->code;
	sErr->level = p->level;
	sErr->line = p->line;

        int cpyLen = 1 + snprintf(sErr->message, GO_ERR_INIT, "%s", p->message);
	if (cpyLen > GO_ERR_INIT) {
		free(sErr->message);
		sErr->message = malloc(cpyLen);
		snprintf(sErr->message, cpyLen, "%s", p->message);
	}

	if (p->node !=NULL) {
		cpyLen = 1 + snprintf(sErr->node, GO_ERR_INIT, "%s", (((xmlNodePtr) p->node)->name));
		if (cpyLen > GO_ERR_INIT) {
			free(sErr->node);
			sErr->node= malloc(cpyLen);
			snprintf(sErr->node, cpyLen, "%s", (((xmlNodePtr) p->node)->name));
		}
	}
}

static struct xsdParserResult cParseUrlSchema(const char *url, const short int options) {
	xmlLineNumbersDefault(1);
	bool err = false;
	struct xsdParserResult parserResult;
	char *errBuf=NULL;
	struct errCtx *ectx=malloc(sizeof(*ectx));
	ectx->errBuf=calloc(GO_ERR_INIT, sizeof(char));

	xmlSchemaPtr schema = NULL;
	xmlSchemaParserCtxtPtr schemaParserCtxt = NULL;

	schemaParserCtxt = xmlSchemaNewParserCtxt(url);

	if (schemaParserCtxt == NULL) {
		err = true;
		strcpy(ectx->errBuf, "Xsd parser internal error");
	}
	else
	{
		if (options & P_ERR_EXT) {
			xmlSchemaSetParserErrors(schemaParserCtxt, noOutputCallback, noOutputCallback, NULL);
			xmlSetGenericErrorFunc(ectx, genErrorCallback);
		} else {
			xmlSetGenericErrorFunc(NULL, noOutputCallback);
			xmlSchemaSetParserErrors(schemaParserCtxt, genErrorCallback, noOutputCallback, ectx);
		}

		schema = xmlSchemaParse(schemaParserCtxt);

		xmlSchemaFreeParserCtxt(schemaParserCtxt);
		if (schema == NULL) {
			err = true;
			char *tmp = malloc(strlen(ectx->errBuf) + 1);
			memcpy(tmp, ectx->errBuf, strlen(ectx->errBuf) + 1);
			free(ectx->errBuf);
			ectx->errBuf = tmp;
		}
	}

	errBuf=malloc(strlen(ectx->errBuf)+1);
	memcpy(errBuf,  ectx->errBuf, strlen(ectx->errBuf)+1);
	free(ectx->errBuf);
	free(ectx);
	parserResult.schemaPtr=schema;
	parserResult.errorStr=errBuf;
	errno = err ? -1 : 0;
	return parserResult;
}

static struct xmlParserResult cParseDoc(const char *goXmlSource, const int goXmlSourceLen, const short int options) {
	xmlLineNumbersDefault(1);
	bool err = false;
	struct xmlParserResult parserResult;
	char *errBuf=NULL;
	struct errCtx *ectx=malloc(sizeof(*ectx));
	ectx->errBuf=calloc(GO_ERR_INIT, sizeof(char));;

	xmlDocPtr doc=NULL;
	xmlParserCtxtPtr xmlParserCtxt=NULL;

	if (goXmlSourceLen == 0) {
		err = true;
		if (options & P_ERR_EXT) {
			strcpy(ectx->errBuf, "parser error : Document is empty");
		} else {
			strcpy(ectx->errBuf, "Malformed xml document");
		}
	} else {
		xmlParserCtxt = xmlNewParserCtxt();

		if (xmlParserCtxt == NULL) {
			err = true;
			strcpy(ectx->errBuf, "Xml parser internal error");
		}
		else
		{
			if (options & P_ERR_EXT) {
				xmlSetGenericErrorFunc(ectx, genErrorCallback);
			} else {
				xmlSetGenericErrorFunc(NULL, noOutputCallback);
			}

			doc = xmlParseMemory(goXmlSource, goXmlSourceLen);

			xmlFreeParserCtxt(xmlParserCtxt);
			if (doc == NULL) {
				err = true;
				if (options & P_ERR_EXT) {
					char *tmp = malloc(strlen(ectx->errBuf) + 1);
					memcpy(tmp, ectx->errBuf, strlen(ectx->errBuf) + 1);
					free(ectx->errBuf);
					ectx->errBuf = tmp;
				} else {
					strcpy(ectx->errBuf, "Malformed xml document");
				}
			}
		}
	}

	errBuf=malloc(strlen(ectx->errBuf)+1);
	memcpy(errBuf,  ectx->errBuf, strlen(ectx->errBuf)+1);
	free(ectx->errBuf);
	free(ectx);
	parserResult.docPtr=doc;
	parserResult.errorStr=errBuf;
	errno = err ? -1 : 0;
	return parserResult;

}

static struct simpleXmlError *cValidate(const xmlDocPtr doc, const xmlSchemaPtr schema) {
	xmlLineNumbersDefault(1);
	bool err = false;
	int schemaErr=0;

	struct simpleXmlError *simpleError = malloc(sizeof(*simpleError));
	simpleError->message = calloc(GO_ERR_INIT, sizeof(char));
	simpleError->node = calloc(GO_ERR_INIT, sizeof(char));

	if (schema == NULL) {
		err = true;
		strcpy(simpleError->message, "Xsd schema null pointer");
	}
	else if (doc == NULL) {
		err = true;
		strcpy(simpleError->message, "Xml schema null pointer");
	}
	else
	{
		xmlSchemaValidCtxtPtr schemaCtxt;
		schemaCtxt = xmlSchemaNewValidCtxt(schema);

		if (schemaCtxt == NULL) {
			err = true;
			strcpy(simpleError->message, "Xml validation internal error");
		}
		else
		{
			xmlSchemaSetValidStructuredErrors(schemaCtxt, simpleStructErrorCallback, simpleError);
			schemaErr = xmlSchemaValidateDoc(schemaCtxt, doc);
			xmlSchemaFreeValidCtxt(schemaCtxt);

			if (schemaErr > 0)
			{
				err = true;
			}
			else if (schemaErr < 0)
			{
				err = true;
				strcpy(simpleError->message, "Xml validation internal error");
			}
		}
	}

	errno = err ? -1 : 0;
	return simpleError;
}


*/
import "C"
import (
	"runtime"
	"strings"
	"time"
	"unsafe"
)

// XsdHandler handles schema parsing and validation and wraps a pointer to libxml2's xmlSchemaPtr.
type XsdHandler struct {
	schemaPtr C.xmlSchemaPtr
}

// XmlHandler handles xml parsing and wraps a pointer to libxml2's xmlDocPtr.
type XmlHandler struct {
	docPtr C.xmlDocPtr
}

// Initializes the libxml2 parser, suggested for multithreading
func libXml2Init() {
	C.init()
}

// Cleans up the libxml2 parser
func libXml2Cleanup() {
	C.cleanup()
}

// The helper function for parsing xml
func parseXmlMem(inXml []byte, options Options) (C.xmlDocPtr, error) {

	strXml := C.CString(string(inXml))
	defer C.free(unsafe.Pointer(strXml))
	pRes, err := C.cParseDoc(strXml, C.int(len(inXml)), C.short(options))

	defer C.free(unsafe.Pointer(pRes.errorStr))
	if err != nil {
		rStr := C.GoString(pRes.errorStr)
		return nil, XmlParserError{errorMessage{strings.Trim(rStr, "\n")}}
	}
	return pRes.docPtr, nil
}

// The helper function for parsing the schema
func parseUrlSchema(url string, options Options) (C.xmlSchemaPtr, error) {
	strUrl := C.CString(url)
	defer C.free(unsafe.Pointer(strUrl))

	pRes, err := C.cParseUrlSchema(strUrl, C.short(options))
	defer C.free(unsafe.Pointer(pRes.errorStr))
	if err != nil {
		rStr := C.GoString(pRes.errorStr)
		return nil, XsdParserError{errorMessage{strings.Trim(rStr, "\n")}}
	}
	return pRes.schemaPtr, nil
}

// Helper function for validating given an xml document
func validateWithXsd(xmlHandler *XmlHandler, xsdHandler *XsdHandler) error {
	sErr, err := C.cValidate(xmlHandler.docPtr, xsdHandler.schemaPtr)
	defer freeSimpleXmlError(sErr)
	if err != nil {
		return ValidationError{
			Code:     int(sErr.code),
			Message:  strings.Trim(C.GoString(sErr.message), "\n"),
			Level:    int(sErr.level),
			Line:     int(sErr.line),
			NodeName: C.GoString(sErr.node),
		}
	}
	return nil
}

// Wrapper for the xmlSchemaFree function
func freeSchemaPtr(xsdHandler *XsdHandler) {
	if xsdHandler.schemaPtr != nil {
		C.xmlSchemaFree(xsdHandler.schemaPtr)
	}
}

// Wrapper for the xmlFreeDoc function
func freeDocPtr(xmlHandler *XmlHandler) {
	if xmlHandler.docPtr != nil {
		C.xmlFreeDoc(xmlHandler.docPtr)
	}
}

// Free C struct
func freeSimpleXmlError(sxe *C.struct_simpleXmlError) {
	C.free(unsafe.Pointer(sxe.message))
	C.free(unsafe.Pointer(sxe.node))
	C.free(unsafe.Pointer(sxe))
}

// Ticker for gc and malloc_trim
func gcTicker(d time.Duration, quit chan struct{}) {
	ticker := time.NewTicker(d)
	for {
		select {
		case <-ticker.C:
			runtime.GC()
			C.malloc_trim(0)
		case <-quit:
			ticker.Stop()
			return
		}
	}
}
