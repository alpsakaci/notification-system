#!/bin/bash

echo "🚀 Başlatılıyor: API Rate Limit Testi (Tekil Endpoint - 500 İstek)"

echo "⏳ 500 adet bildirim paralel olarak API'ye gönderiliyor..."

# Geçici bir dosya oluştur
TMP_FILE=$(mktemp)

# 500 isteği asenkron (paralel) olarak gönder
for i in {1..500}; do
  PAYLOAD="{\"recipient\": \"user$i@example.com\", \"channel\": \"email\", \"content\": \"Rate Limit Test $i\", \"priority\": \"normal\"}"
  
  (
    HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8080/api/v1/notifications \
      -H "Content-Type: application/json" \
      -d "$PAYLOAD")
    echo "$HTTP_STATUS" >> "$TMP_FILE"
  ) &
done

# Tüm arkaplan işlemlerinin bitmesini bekle
wait

echo "✅ Tüm istekler gönderildi."
echo "📊 Sonuçlar:"

# Başarılı ve Rate Limit'e takılanları say
SUCCESS_COUNT=$(grep -c "202" "$TMP_FILE")
TOO_MANY_REQUESTS_COUNT=$(grep -c "429" "$TMP_FILE")

echo "✅ Başarılı (202 Accepted): $SUCCESS_COUNT"
echo "🚫 Rate Limit'e Takılan (429 Too Many Requests): $TOO_MANY_REQUESTS_COUNT"

if [ "$TOO_MANY_REQUESTS_COUNT" -gt 0 ]; then
    echo "🎉 Rate limiting API katmanında başarıyla çalışıyor!"
else
    echo "❌ Rate limit tetiklenmedi. Belki makineniz saniyede 100 paralel isteği atabilecek kadar hızlı değildir veya limit yüksektir."
fi

# Geçici dosyayı sil
rm -f "$TMP_FILE"
