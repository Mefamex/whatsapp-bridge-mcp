# =============================================================================
# WhatsApp Bridge & MCP — Makefile
#
# Kendi kendini belgeleyen (Self-documenting) görev yöneticisi.
# Terminalde sadece 'make' yazıldığında kullanılabilir komutları listeler.
#
# ZORUNLU YAŞAM DÖNGÜSÜ HEDEFLERİ:
# [install, reinstall, update, run, clean, remove]
# =============================================================================

.DEFAULT_GOAL := help
.PHONY: help install reinstall update run dev lint test verify build clean remove run-mcp

help: ## Mevcut komutları ve açıklamalarını listeler
	@echo "Kullanım: make [komut]"
	@echo ""
	@echo "Kullanılabilir Komutlar:"
	@LC_ALL=C grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'

# --- 1. KURULUM VE YAŞAM DÖNGÜSÜ ---

install: ## Bağımlılıkları kurar (Go modülleri + Python uv sanal ortamı)
	@echo "==> Go bağımlılıkları indiriliyor..."
	cd bridge && go mod download
	@echo "==> Python sanal ortamı ve MCP paketleri kuruluyor..."
	uv venv .venv
	uv pip install -r mcp_server/requirements.txt
	@echo "==> Kurulum tamamlandı."

reinstall: remove install ## Kurulu ortamı kaldırıp sıfırdan yeniden kurar
	@echo "==> Temiz yeniden kurulum tamamlandı."

update: ## Bağımlılıkları en son sürümlere günceller
	@echo "==> Bağımlılıklar güncelleniyor..."
	cd bridge && go get -u ./... && go mod tidy
	uv pip install --upgrade -r mcp_server/requirements.txt

remove: ## Bağımlılıkları, sanal ortamı ve derleme çıktılarını kazır
	@echo "==> Sanal ortam ve derleme çıktıları kaldırılıyor..."
	rm -rf .venv bin/ bridge/whatsapp-bridge
	@echo "Kaldırma işlemi tamamlandı."

# --- 2. ÇALIŞTIRMA VE GELİŞTİRME ---

build: ## Go WhatsApp Bridge ikili dosyasını derler (bin/whatsapp-bridge)
	@echo "==> WhatsApp Bridge derleniyor..."
	mkdir -p bin
	cd bridge && go build -o ../bin/whatsapp-bridge .
	@echo "==> Derleme başarılı: bin/whatsapp-bridge"

run: build ## WhatsApp Bridge daemon'ını çalıştırır
	@echo "==> WhatsApp Bridge başlatılıyor (Port 31415)..."
	./bin/whatsapp-bridge

run-mcp: ## WhatsApp MCP Server'ı STDIO modunda çalıştırır
	@echo "==> WhatsApp MCP Server başlatılıyor..."
	.venv/bin/python -m mcp_server

dev: build ## Köprüyü geliştirme modunda çalıştırır
	@echo "==> Geliştirme modu başlatılıyor..."
	./bin/whatsapp-bridge

# --- 3. DENETİM VE DOĞRULAMA ---

lint: ## Go ve Python için statik kod analizi yapar
	@echo "==> Go vet analizi yapılıyor..."
	cd bridge && go vet ./...
	@echo "==> Python sözdizimi denetleniyor..."
	.venv/bin/python -m py_compile mcp_server/*.py mcp_server/**/*.py

test: ## Go ve Python birim testlerini çalıştırır
	@echo "==> Go birim testleri koşturuluyor..."
	cd bridge && go test ./...
	@echo "==> Python MCP testleri koşturuluyor..."
	.venv/bin/python -m unittest discover -s tests -p 'test_*.py' -v

verify: lint test ## Commit/Push öncesi tam doğrulama zincirini çalıştırır (lint + test)
	@echo "==> TÜM DOĞRULAMALAR BAŞARIYLA GEÇTİ (ALL_OK)."

# --- 4. TEMİZLİK ---

clean: ## Geçici dosyaları, test önbelleklerini ve kalıntıları temizler
	@echo "==> Kalıntılar temizleniyor..."
	find . -type d -name "__pycache__" -exec rm -rf {} + 2>/dev/null || true
	rm -rf .pytest_cache .ruff_cache tests/__pycache__ mcp_server/**/__pycache__
	@echo "Temizlik tamamlandı."
