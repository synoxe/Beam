# Beam

Beam, Go ile yazılmış, aynı ağdaki cihazlar arasında dosya paylaşımı yapmak için geliştirilmiş bir CLI aracıdır.

## Özellikler

- TCP üzerinden dosya gönderme ve alma
- UDP broadcast ile ağdaki Beam cihazlarını keşfetme
- SHA-256 checksum doğrulaması
- Geçici `.part` dosyası ile güvenli kayıt mantığı
- Güvenli dosya adı temizleme
- Aynı isimde dosya varsa otomatik yeniden adlandırma
- Gönderim ve alım sırasında progress bar
- Anlık aktarım hızı ve tahmini kalan süre

## Proje Yapısı

```text
beam/
├── cmd/
│   └── beam/
│       └── main.go
├── internal/
│   ├── commands/
│   │   └── root.go
│   ├── discovery/
│   │   └── discovery.go
│   ├── hashutil/
│   │   └── hash.go
│   ├── progress/
│   │   └── progress.go
│   ├── protocol/
│   │   └── messages.go
│   ├── storage/
│   │   └── files.go
│   └── transfer/
│       ├── receiver.go
│       └── sender.go
└── go.mod
```

## Gereksinimler

- Go 1.22 veya üstü önerilir
- Aynı yerel ağda en az iki cihaz
- Firewall kuralları TCP paylaşım portuna ve UDP discovery portuna izin vermelidir

## Kurulum

### Geliştirme sırasında çalıştırma

```bash
go run ./cmd/beam help
```

### Binary üretme

```bash
go build -o beam.exe ./cmd/beam
```

Windows PowerShell:

```powershell
.\beam.exe help
```

### Her yerden `beam` komutunu çalıştırmak

```bash
go install ./cmd/beam
```

Bu komut `beam.exe` dosyasını genelde şu klasöre koyar:

```text
C:\Users\KULLANICI_ADI\go\bin
```

Bu klasörü PATH değişkenine eklersen şu komutları doğrudan kullanabilirsin:

```bash
beam help
beam discover
beam receive --port 9000
beam send --file test.txt --to 192.168.1.25:9000
```

## Kullanım

### Yardım ekranı

```bash
beam help
```

### Dosya almak

```bash
beam receive --port 9000 --dir downloads
```

Açıklama:
- `--port`: TCP dinleme portu
- `--dir`: gelen dosyaların kaydedileceği klasör

Varsayılanlar:
- port: `9000`
- klasör: `downloads`

### Ağdaki Beam cihazlarını bulmak

```bash
beam discover
```

Bu komut yerel ağda discovery broadcast gönderir ve cevap veren Beam cihazlarını listeler.

Örnek çıktı:

```text
Ağda Beam cihazları aranıyor...
Bulunan cihazlar:
1. arif-laptop - 192.168.1.25:9000
2. desktop-pc - 192.168.1.33:9000
```

### Dosya göndermek

```bash
beam send --file ./test.txt --to 192.168.1.25:9000
```

Açıklama:
- `--file`: gönderilecek dosya yolu
- `--to`: hedef cihazın `ip:port` bilgisi

## Örnek Akış

### 1. Alıcı cihazda

```bash
beam receive --port 9000 --dir downloads
```

### 2. Gönderici cihazda cihazları tara

```bash
beam discover
```

### 3. Gönderici cihazda dosyayı yolla

```bash
beam send --file ./rapor.pdf --to 192.168.1.25:9000
```

## Progress Bar

Beam gönderim ve alım sırasında terminalde ilerleme çubuğu gösterir.

Örnek:

```text
Sending   [===========                 ]  41.32%  20.66 MB / 50.00 MB  4.92 MB/s  ETA: 6s
Receiving [===========                 ]  41.32%  20.66 MB / 50.00 MB  4.87 MB/s  ETA: 6s
```

Gösterilen bilgiler:
- yüzde ilerleme
- aktarılan veri / toplam veri
- anlık aktarım hızı
- tahmini kalan süre

## Nasıl Çalışır

### Dosya Transferi

Beam, dosya aktarımı için TCP kullanır.

Akış:
1. Sender hedef cihaza TCP ile bağlanır
2. Önce dosya metadata bilgisi gönderilir
3. Receiver metadata'yı doğrular
4. Receiver `ACCEPT` mesajı döner
5. Sender dosya byte'larını gönderir
6. Receiver dosyayı `.part` uzantılı geçici dosyaya yazar
7. Receiver SHA-256 checksum kontrolü yapar
8. Doğrulama başarılıysa dosya gerçek ismine taşınır
9. Receiver `DONE` mesajı gönderir

### Discovery

Beam, cihaz keşfi için UDP broadcast kullanır.

Akış:
1. `beam discover` komutu `BEAM_DISCOVER` mesajı yollar
2. Receive modunda çalışan Beam cihazları bu mesajı dinler
3. Her cihaz kendi `name`, `ip`, `port` bilgisini geri yollar
4. Discover komutu gelen cevapları listeler

## Güvenlik ve Sağlamlık

Bu sürümde şu önlemler vardır:

- Dosya adı temizleme
- Path traversal denemelerini engelleme
- Aynı isimde dosya varsa yeni isim üretme
- `.part` dosyası ile yarım transferleri ayırma
- SHA-256 checksum doğrulama

Şu özellikler henüz yoktur:

- TLS
- token tabanlı kimlik doğrulama
- resume desteği
- klasör gönderme
- çoklu dosya gönderimi

## Hata Notları

### Discovery çalışmıyorsa

Bazı ağlarda UDP broadcast engellenebilir. Özellikle:
- misafir ağları
- kurumsal ağlar
- firewall aktif sistemler

Bu durumda doğrudan IP ile gönderim yapabilirsin:

```bash
beam send --file ./test.txt --to 192.168.1.25:9000
```

### Port kullanımda hatası

Aynı portu kullanan başka bir uygulama olabilir. Farklı port dene:

```bash
beam receive --port 9100 --dir downloads
```

### Dosya alınmadıysa

Şunları kontrol et:
- receiver açık mı
- IP doğru mu
- port doğru mu
- firewall izin veriyor mu
- dosya yolu gerçekten var mı

## Geliştirme

Projeyi çalıştırmak:

```bash
go run ./cmd/beam help
```

Derlemek:

```bash
go build -o beam.exe ./cmd/beam
```

Kurmak:

```bash
go install ./cmd/beam
```

## Yol Haritası

Planlanan geliştirmeler:

- token ile doğrulama
- TLS desteği
- aktarım geçmişi
- resume desteği
- klasör gönderme
- çoklu dosya gönderme
- daha gelişmiş progress görünümü

## Lisans

Bu proje eğitim ve geliştirme amaçlı hazırlanmıştır.
