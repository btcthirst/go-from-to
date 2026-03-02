Code Review: internal/ui

internal/ui/app.go
⚠️ OnTransactionsChanged перезаписується
NewTransactionsScreen присвоює state.OnTransactionsChanged напряму (рядок у transactions_screen.go). Якщо newReport311Section також підписується через ланцюжок (prevOnChanged), це крихко — порядок ініціалізації важливий, і додавання ще одного підписника зламає ланцюг. Рекомендація: замінити на slice колбеків або sync/event bus.
go// Замість одного func()
type AppState struct {
    txListeners []func()
}
func (s *AppState) AddTransactionListener(fn func()) {
    s.txListeners = append(s.txListeners, fn)
}
func (s *AppState) NotifyTransactionsChanged() {
    for _, fn := range s.txListeners { fn() }
}
✅ sync.RWMutex використовується коректно у AppendTransactions / GetTransactions / ClearTransactions.
⚠️ fyne.Do() викликається з головного потоку при Run() — ризику немає зараз, але Run() запускається в main goroutine, тож будь-який fyne.Do в ініціалізації зависне. Поки не проблема, але варто документувати.

internal/ui/import_screen.go
🔴 Race condition на files slice
files не захищений при читанні в fileList callbacks:
go// fileList.UpdateItem читає files[id] без блокування
func(id widget.ListItemID, item fyne.CanvasObject) {
    f := files[id]  // ← читання без mu.RLock
fileList.Refresh() викликається з fyne.Do() (головний потік), але між перевіркою id >= len(files) і files[id] може статись запис з іншої горутини якщо importFile викликається паралельно. Виправлення — взяти знімок:
gofunc(id widget.ListItemID, item fyne.CanvasObject) {
    mu.RLock()
    if id >= len(files) { mu.RUnlock(); return }
    f := files[id]
    mu.RUnlock()
    // ...
}
⚠️ win.SetOnDropped перезаписується при кожному виклику NewImportScreen
Якщо екран перестворюється, попередній обробник губиться. Зараз екрани не перестворюються, але це тихий ризик.
⚠️ reader.Close() після reader.URI().Path()
У dialog.NewFileOpen колбеку:
gopath := reader.URI().Path()
reader.Close()  // ✅ — ок, але варто defer
Якщо між Path() і Close() паніка — ресурс витікає. Краще defer reader.Close().
✅ Парсинг і категоризація в горутині, UI-оновлення через fyne.Do() — правильно.
✅ Помилка детекції обробляється явно (відображається через dialog.ShowError).

internal/ui/transactions_screen.go
🔴 state.OnTransactionsChanged перезаписує попередній колбек
gostate.OnTransactionsChanged = func() {  // ← безумовне присвоєння
    refreshCache()
    table.Refresh()
}
report_screen.go → newReport311Section робить те саме через prevOnChanged. Якщо NewTransactionsScreen викликається після NewReportScreen, колбек 311-звіту губиться. Рішення — slice (див. вище).
⚠️ filtered() звертається до state.Transactions без блокування
gofunc filtered(state *AppState, ...) []*models.Transaction {
    for _, tx := range state.Transactions {  // ← немає RLock
Треба state.GetTransactions() замість прямого доступу:
gofunc filtered(state *AppState, search, category, txType string) []*models.Transaction {
    txs := state.GetTransactions()
    for _, tx := range txs {
✅ cachedTxs оновлюється в refreshCache() який викликається з головного потоку — race між refreshCache і рендером таблиці відсутній.
⚠️ tableHeaders[id.Col] — немає перевірки id.Col >= len(tableHeaders)
При colCount != len(tableHeaders) (наприклад, хтось додав колонку в константи але не в slice) — паніка. Додати:
goif id.Col >= len(tableHeaders) { lbl.SetTexts("?", ""); return }
✅ Перевірка id.Row-1 >= len(cachedTxs) присутня — захист від index out of range є.

internal/ui/report_screen.go
⚠️ buildReport — бізнес-логіка в UI шарі
Функція buildReport (підрахунок TotalIncome, TotalExpense, ByCategory, ByMonth) знаходиться в report_screen.go, а не в reports/ або models/. Це порушує розподіл відповідальності і ускладнює тестування. Рекомендація — перенести в internal/reports/builder.go.
⚠️ prevOnChanged ланцюжок у newReport311Section
goprevOnChanged := state.OnTransactionsChanged
state.OnTransactionsChanged = func() {
    if prevOnChanged != nil { prevOnChanged() }
    refreshPreview311()
}
Це ланцюгування хибне якщо NewTransactionsScreen викликається після — воно перезапише весь ланцюг. Проблема системна (див. вище).
✅ runInBackground коректно блокує кнопки та оновлює UI через fyne.Do().
✅ filterByDate і filterByMonth — чисті функції без side effects.
⚠️ writer.Close() перед генерацією звіту
gooutputPath := writer.URI().Path()
writer.Close()  // ← файл закривається до запису
На деяких ОС (Windows) файл заблокований після NewFileSave до явного закриття. Закриття до запису — правильно. ✅ Але якщо writer — це вже відкритий файловий дескриптор, а outputPath передається в excelize.SaveAs — тоді все ок, excelize відкриває файл сам. Варто перевірити поведінку на Windows.

internal/ui/categories_screen.go
⚠️ categories — локальна копія, не синхронізована зі state.Config
gocategories := state.Config.CategoryNames()
// ...
state.Config.AddCategory(name)
categories = state.Config.CategoryNames()  // ← оновлюється вручну
Якщо CategoryNames() повертає новий slice при кожному виклику — ок. Але при видаленні категорії:
gostate.Config.RemoveCategory(selectedCategory)
categories = state.Config.CategoryNames()  // ← оновлюється ✅
Загалом коректно, але крихко — легко забути оновити при додаванні нового шляху зміни.
⚠️ countTxInCategory звертається до state.Transactions без блокування
gofunc countTxInCategory(state *AppState, category string) int {
    for _, tx := range state.Transactions {  // ← немає RLock
Аналогічно до filtered() — замінити на state.GetTransactions().
✅ Діалог підтвердження перед видаленням категорії — є.
✅ Транзакції переводяться в "Uncategorized" при видаленні категорії — є.

internal/ui/widgets/
tooltip_label.go
⚠️ timer goroutine може спрацювати після знищення віджета
gol.timer = time.AfterFunc(tooltipDelay, func() {
    fyne.Do(func() { l.showTooltip(pos) })
})
Якщо таблиця оновилась і клітинка перестворена, старий TooltipLabel може показати tooltip. Некритично (popup зникне при наступному MouseOut), але виглядає як артефакт.
✅ sync.Mutex для timer і popup — коректно.
✅ fyne.Do() в таймері — правильно.
file_droper.go
✅ DroppedFiles викликається з Window.SetOnDropped з перевіркою координат — нестандартне, але працююче рішення.
⚠️ Назва файлу file_droper.go — опечатка (одне p). Не функціональна проблема, але варто виправити для консистентності.
autocomlete.go
⚠️ Назва файлу autocomlete.go — опечатка (одне p). Аналогічно.
⚠️ e.filtered = e.filtered[:0] — зрізання slice без звільнення пам'яті. При великому AllOptions може тримати зайву пам'ять. Для малих довідників (провайдери) — не критично.

Загальний висновок
КатегоріяСтатусRace conditions🔴 files в import_screen, state.Transactions в filtered/countTxOnTransactionsChanged архітектура🔴 Системна проблема, перезапис колбеківБізнес-логіка в UI⚠️ buildReport в report_screen.goЗахист індексів таблиці⚠️ tableHeaders[id.Col] без перевіркиОпечатки в назвах файлів⚠️ file_droper.go, autocomlete.go
Обов'язкові зміни перед production:

1. Замінити state.OnTransactionsChanged func() на []func() з методом AddTransactionListener.
2. Замінити прямий доступ state.Transactions у filtered() і countTxInCategory() на state.GetTransactions().
3. Додати mu.RLock/RUnlock в callbacks fileList в import_screen.go.
4. Перенести buildReport в internal/reports/builder.go.

Виправлення пунктів 1–3 блокують production-запуск. Пункт 4 — важливий для підтримуваності, але не критичний для стабільності.