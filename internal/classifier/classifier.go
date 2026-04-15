package classifier

import (
	"strings"

	"paramind/internal/model"
)

var keyNormalizer = strings.NewReplacer(
	"_", "",
	"-", "",
	"\u00e1", "a",
	"\u00e0", "a",
	"\u00e2", "a",
	"\u00e3", "a",
	"\u00e4", "a",
	"\u00e9", "e",
	"\u00e8", "e",
	"\u00ea", "e",
	"\u00eb", "e",
	"\u00ed", "i",
	"\u00ec", "i",
	"\u00ee", "i",
	"\u00ef", "i",
	"\u00f3", "o",
	"\u00f2", "o",
	"\u00f4", "o",
	"\u00f5", "o",
	"\u00f6", "o",
	"\u00fa", "u",
	"\u00f9", "u",
	"\u00fb", "u",
	"\u00fc", "u",
	"\u00e7", "c",
	"\u00f1", "n",
)

type categoryDefinition struct {
	Name              string
	Priority          int
	ExactKeys         []string
	PartialKeys       []string
	Hypotheses        []string
	exactSet          map[string]struct{}
	normalizedExact   map[string]struct{}
	normalizedPartial []string
}

type Classifier struct {
	categories []categoryDefinition
}

func New() *Classifier {
	definitions := defaultDefinitions()
	for i := range definitions {
		definitions[i].exactSet = make(map[string]struct{}, len(definitions[i].ExactKeys))
		definitions[i].normalizedExact = make(map[string]struct{}, len(definitions[i].ExactKeys))
		definitions[i].normalizedPartial = make([]string, 0, len(definitions[i].PartialKeys))

		for _, key := range definitions[i].ExactKeys {
			lower := strings.ToLower(key)
			definitions[i].exactSet[lower] = struct{}{}
			definitions[i].normalizedExact[normalizeKey(lower)] = struct{}{}
		}

		for _, key := range definitions[i].PartialKeys {
			definitions[i].normalizedPartial = append(definitions[i].normalizedPartial, normalizeKey(key))
		}
	}

	return &Classifier{categories: definitions}
}

func KnownCategories() map[string]struct{} {
	known := make(map[string]struct{}, len(defaultDefinitions()))
	for _, category := range defaultDefinitions() {
		known[category.Name] = struct{}{}
	}
	return known
}

func (c *Classifier) Classify(param, value string) (model.Finding, bool) {
	paramLower := strings.ToLower(strings.TrimSpace(param))
	paramNormalized := normalizeKey(paramLower)

	bestIndex := -1
	bestConfidence := model.ConfidenceNone

	for i := range c.categories {
		confidence := c.categories[i].match(paramLower, paramNormalized)
		if confidence == model.ConfidenceNone {
			continue
		}

		if bestIndex == -1 || c.categories[i].Priority < c.categories[bestIndex].Priority {
			bestIndex = i
			bestConfidence = confidence
			continue
		}

		if c.categories[i].Priority == c.categories[bestIndex].Priority && confidenceRank(confidence) > confidenceRank(bestConfidence) {
			bestIndex = i
			bestConfidence = confidence
		}
	}

	if bestIndex == -1 {
		return model.Finding{}, false
	}

	hypotheses := append([]string(nil), c.categories[bestIndex].Hypotheses...)

	return model.Finding{
		Param:      param,
		Value:      value,
		Class:      c.categories[bestIndex].Name,
		Confidence: bestConfidence,
		Hypotheses: hypotheses,
	}, true
}

func (c categoryDefinition) match(paramLower, paramNormalized string) model.Confidence {
	if paramLower == "" {
		return model.ConfidenceNone
	}

	if _, exact := c.exactSet[paramLower]; exact {
		return model.ConfidenceHigh
	}

	if _, normalized := c.normalizedExact[paramNormalized]; normalized {
		return model.ConfidenceMedium
	}

	for _, candidate := range c.normalizedPartial {
		if candidate != "" && strings.Contains(paramNormalized, candidate) {
			return model.ConfidenceLow
		}
	}

	return model.ConfidenceNone
}

func normalizeKey(value string) string {
	return keyNormalizer.Replace(strings.ToLower(strings.TrimSpace(value)))
}

func confidenceRank(value model.Confidence) int {
	switch value {
	case model.ConfidenceHigh:
		return 3
	case model.ConfidenceMedium:
		return 2
	case model.ConfidenceLow:
		return 1
	default:
		return 0
	}
}

func defaultDefinitions() []categoryDefinition {
	return []categoryDefinition{
		{
			Name:     "auth",
			Priority: 1,
			ExactKeys: []string{
				"token", "access_token", "auth", "session", "jwt", "api_key",
				"apikey", "secret", "password", "passwd", "key",
				"senha", "sessao", "segredo", "chave", "token_acesso",
				"chave_api", "senha_api", "contrasena", "sesion", "clave",
				"clave_api", "secreto",
			},
			PartialKeys: []string{
				"auth", "token", "session", "key", "secret",
				"senha", "sessao", "segredo", "chave", "contrasena",
				"sesion", "clave",
			},
			Hypotheses: []string{
				"token_leakage",
				"session_fixation",
				"account_takeover_vector",
				"weak_auth_flow",
			},
		},
		{
			Name:     "redirect",
			Priority: 2,
			ExactKeys: []string{
				"url", "target", "dest", "destination", "redirect",
				"redirect_url", "redirect_uri", "return",
				"return_url", "return_to", "next", "continue",
				"goto", "callback", "to",
				"destino", "retorno", "url_retorno", "retorno_url",
				"prox", "proximo", "proxima", "continuar", "voltar",
				"voltar_para", "redir", "redirecionar", "destino_url",
				"url_destino", "siguiente", "siguiente_url", "url_siguiente",
			},
			PartialKeys: []string{
				"redirect", "return", "next", "callback",
				"redir", "retorno", "destino", "proxim", "siguient",
				"voltar",
			},
			Hypotheses: []string{
				"open_redirect",
				"redirect_validation_bypass",
				"oauth_redirect_abuse",
				"token_leak_via_redirect",
			},
		},
		{
			Name:     "ssrf",
			Priority: 3,
			ExactKeys: []string{
				"uri", "link", "src", "source", "feed", "host", "domain",
				"proxy", "fetch", "site", "webhook", "endpoint", "api", "image_url",
				"origem", "origen", "fonte", "fuente", "servidor", "dominio",
				"endereco", "direccion", "url_imagem", "imagem_url",
				"url_imagen", "imagen_url", "webhook_url", "url_webhook",
				"recurso", "remoto",
			},
			PartialKeys: []string{
				"url", "uri", "fetch", "host", "domain", "webhook",
				"origem", "origen", "fonte", "fuente", "servidor",
				"dominio", "endereco", "direccion", "imagem", "imagen",
			},
			Hypotheses: []string{
				"ssrf",
				"blind_ssrf",
				"open_proxy",
				"internal_host_access",
			},
		},
		{
			Name:     "file",
			Priority: 4,
			ExactKeys: []string{
				"file", "filepath", "path", "folder", "dir", "document",
				"download", "template", "view", "page", "include", "layout",
				"arquivo", "caminho", "diretorio", "diretorio_base",
				"pasta", "anexo", "plantilla", "vista", "descarga",
				"ruta", "ruta_archivo",
			},
			PartialKeys: []string{
				"file", "path", "dir", "template", "view",
				"arquivo", "caminho", "diretor", "pasta", "anexo",
				"plantilla", "vista", "ruta",
			},
			Hypotheses: []string{
				"lfi",
				"path_traversal",
				"arbitrary_file_read",
			},
		},
		{
			Name:     "id",
			Priority: 5,
			ExactKeys: []string{
				"id", "user", "user_id", "userid", "account",
				"account_id", "profile", "order", "order_id",
				"doc", "doc_id", "record", "item", "product",
				"invoice", "ticket", "message", "ref", "reference",
				"usuario", "usuario_id", "conta", "conta_id", "pedido",
				"pedido_id", "cliente", "cliente_id", "documento",
				"documento_id", "registro", "registro_id", "produto",
				"produto_id", "fatura", "fatura_id", "mensagem",
				"mensagem_id", "referencia", "referencia_id", "perfil",
				"perfil_id", "cuenta", "cuenta_id",
			},
			PartialKeys: []string{
				"id", "user", "account", "order", "doc",
				"usuario", "conta", "pedido", "cliente", "document",
				"registro", "produto", "referenc", "perfil", "fatura",
				"cuenta",
			},
			Hypotheses: []string{
				"idor",
				"enumeration",
				"broken_access_control",
			},
		},
		{
			Name:     "sqli",
			Priority: 6,
			ExactKeys: []string{
				"query", "search", "q", "filter", "sort", "order",
				"column", "table", "where",
				"busca", "pesquisa", "buscar", "consulta", "filtro",
				"filtros", "ordem", "ordenacao", "ordenar", "orden",
				"criterio", "coluna", "tabela", "onde",
			},
			PartialKeys: []string{
				"query", "search", "filter", "sort", "order",
				"busca", "pesquisa", "consulta", "filtro", "ordem",
				"orden", "criterio", "coluna", "tabela",
			},
			Hypotheses: []string{
				"sqli",
				"nosqli",
				"query_manipulation",
			},
		},
		{
			Name:     "xss",
			Priority: 7,
			ExactKeys: []string{
				"q", "search", "keyword", "name", "title",
				"comment", "message", "msg", "error", "text",
				"nome", "titulo", "comentario", "comentarios",
				"mensagem", "texto", "erro", "descricao", "termo",
				"palavra_chave", "palavrachave",
			},
			PartialKeys: []string{
				"search", "text", "msg", "name",
				"texto", "mensag", "coment", "nome", "titulo",
				"erro", "descric", "termo",
			},
			Hypotheses: []string{
				"reflected_xss",
				"stored_xss",
			},
		},
		{
			Name:     "debug",
			Priority: 8,
			ExactKeys: []string{
				"debug", "test", "env", "admin", "internal",
				"preview", "render", "template",
				"teste", "ambiente", "interno", "previa",
				"previsualizacao", "rascunho", "homologacao", "homolog",
				"staging", "administracao", "modo_teste", "modo_debug",
				"depuracao", "depuracion",
			},
			PartialKeys: []string{
				"debug", "test", "admin",
				"teste", "interno", "homolog", "staging", "depur",
				"previa",
			},
			Hypotheses: []string{
				"debug_exposure",
				"hidden_feature_abuse",
				"ssti",
			},
		},
	}
}
