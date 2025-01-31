# Lecture Ideas

## **Prednáška: Inside Go Runtime – Garbage Collection, Scheduling, and Optimizations**  

### **Motivácia (Prečo je táto téma dôležitá?)**  
Go je známe svojím efektívnym runtime, ktorý automaticky spravuje pamäť, plánuje goroutines a optimalizuje výkon. No pri škálovateľných a výkonných aplikáciách môže nesprávne pochopenie fungovania runtime viesť k:  
- **Nečakaným latenciám** spôsobeným garbage collectorom (GC)  
- **Zlému využitiu CPU** kvôli nesprávnemu plánovaniu goroutines  
- **Neefektívnej alokácii pamäte**, čo môže viesť k OOM (Out of Memory) chybám  

Cieľom prednášky je vysvetliť, ako Go runtime funguje pod kapotou a ako môžeme jeho správanie optimalizovať.

---

### **Definovanie problémov**  

#### **1. Garbage Collection (GC) – Ako to funguje?**  
- **Stop-the-world momenty vs. concurrent GC**  
- Ako sa líši GC v Go oproti JVM alebo Pythonu  
- Ako identifikovať a optimalizovať vysoký GC overhead  

**Príklad problému:**  
Máš Go aplikáciu s veľkým množstvom krátkodobých objektov, ale performance trpí kvôli nadmernému zberu odpadu.  

---

#### **2. Scheduler a Goroutines – Ako sú plánované?**  
- Ako funguje **M:N scheduling** v Go (goroutines vs. OS threads)  
- **Work stealing a thread parking**  
- Problémy ako **goroutine leaks** alebo **thread starvation**  

**Príklad problému:**  
Aplikácia má tisíce goroutines, ale CPU je stále nevyužité – kde je problém?  

---

#### **3. Optimalizácie pamäte a CPU využitia**  
- Zero-allocation patterns: Ako eliminovať zbytočné alokácie  
- `sync.Pool` – efektívne znovupoužitie objektov  
- `pprof` a `trace` – ako analyzovať výkon aplikácie  

**Príklad problému:**  
Tvoja aplikácia generuje veľa krátkodobých alokácií a GC trávi príliš veľa času ich uvoľňovaním.  

---

### **Riešenia a best practices**  
- Ako efektívne pracovať s GC (`GOGC`, profilovanie, refactoring alokácií)  
- Správne plánovanie goroutines (`runtime.Gosched()`, `sync.WaitGroup`)  
- Ako optimalizovať výkon (`pprof`, `trace`, `allocations`)  

---

### **Sumarizácia a takeaway**  
✅ **Pochopenie GC a jeho dopadu na aplikáciu**  
✅ **Zlepšenie CPU a pamäťovej efektivity aplikácií v Go**  
✅ **Schopnosť analyzovať runtime správanie pomocou nástrojov ako `pprof` a `trace`**  

