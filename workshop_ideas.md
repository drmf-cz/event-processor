## **Workshop: Building a Real-Time Event Processing System in Go**  

### **Motivácia (Prečo je táto téma dôležitá?)**  
Mnohé moderné aplikácie potrebujú **spracovávať udalosti v reálnom čase** (logovanie, alerty, streaming dát). Typické riešenia používajú Kafka, NATS alebo Redis Streams, ale:  
- Ako správne spracovávať veľké množstvo eventov v Go?  
- Ako efektívne paralelizovať prácu s eventmi?  
- Ako zabezpečiť odolnosť voči výpadkom?  

---

### **Definovanie problému**  
Predstavme si systém na spracovanie finančných transakcií v reálnom čase:  
- Eventy prichádzajú cez **Kafka/NATS stream**  
- Potrebujeme **agregovať a filtrovať dáta**  
- Musíme zabezpečiť **presnosť a škálovateľnosť**  

Problémy:  
- Ak sa proces zastaví, nestratíme eventy?  
- Ako zaistiť, že spracujeme každý event len raz?  
- Ako efektívne paralelizovať prácu, aby sme využili všetky CPU jadrá?  

---

### **Riešenie (Čo sa bude budovať v rámci workshopu?)**  
**1. Príprava prostredia**  
- Nastavenie Kafka alebo NATS streamu  
- Implementácia Go konzumenta eventov  

**2. Paralelizované spracovanie eventov**  
- Použitie **worker pool pattern** na efektívne využitie CPU  
- Zabezpečenie **back-pressure** mechanizmu (Rate limiting)  

**3. Idempotencia a zaručenie spracovania**  
- Implementácia **at-least-once processing**  
- Práca s deduplikačnými mechanizmami (napr. Redis na ukladanie ID eventov)  

**4. Logging, monitoring a alerty**  
- Pridanie **OpenTelemetry** na tracing  
- Nastavenie Prometheus + Grafana na metriky  

---

### **Sumarizácia a takeaway**  
✅ **Postavenie základného real-time event processing systému v Go**  
✅ **Použitie worker pool pattern na efektívne paralelné spracovanie**  
✅ **Pridanie observability pomocou tracingu a metriky**  
✅ **Riešenie problémov ako duplikácia eventov a back-pressure**  

---
