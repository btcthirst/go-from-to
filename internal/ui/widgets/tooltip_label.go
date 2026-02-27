package widgets

import (
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const tooltipDelay = 600 * time.Millisecond

// TooltipLabel — лейбл який показує повний текст у popup при наведенні миші
// з затримкою tooltipDelay. Використовується в таблиці для обрізаних клітинок.
//
// Реалізує desktop.Hoverable для отримання MouseIn/MouseOut подій.
type TooltipLabel struct {
	widget.BaseWidget

	Text  string // відображуваний (обрізаний) текст
	Full  string // повний текст для tooltip; якщо "" або == Text — не показується
	Align fyne.TextAlign
	Style fyne.TextStyle

	mu      sync.Mutex
	timer   *time.Timer
	popup   *widget.PopUp
	lastPos fyne.Position
}

// Перевірка що TooltipLabel реалізує desktop.Hoverable.
var _ desktop.Hoverable = (*TooltipLabel)(nil)

func NewTooltipLabel(display, full string) *TooltipLabel {
	l := &TooltipLabel{Text: display, Full: full}
	l.ExtendBaseWidget(l)
	return l
}

// SetTexts оновлює обидва тексти і рефрешує віджет.
func (l *TooltipLabel) SetTexts(display, full string) {
	l.hideTooltip()
	l.Text = display
	l.Full = full
	l.Refresh()
}

// --- desktop.Hoverable ---

func (l *TooltipLabel) MouseIn(ev *desktop.MouseEvent) {
	if l.Full == "" || l.Full == l.Text {
		return
	}
	l.mu.Lock()
	l.lastPos = ev.AbsolutePosition
	if l.timer != nil {
		l.timer.Stop()
	}
	l.timer = time.AfterFunc(tooltipDelay, func() {
		fyne.Do(func() {
			l.mu.Lock()
			pos := l.lastPos
			l.mu.Unlock()
			l.showTooltip(pos)
		})
	})
	l.mu.Unlock()
}

func (l *TooltipLabel) MouseMoved(ev *desktop.MouseEvent) {
	l.mu.Lock()
	l.lastPos = ev.AbsolutePosition
	l.mu.Unlock()
}

func (l *TooltipLabel) MouseOut() {
	l.mu.Lock()
	if l.timer != nil {
		l.timer.Stop()
		l.timer = nil
	}
	l.mu.Unlock()
	fyne.Do(l.hideTooltip)
}

// --- Popup ---

func (l *TooltipLabel) showTooltip(pos fyne.Position) {
	l.hideTooltip()

	wins := fyne.CurrentApp().Driver().AllWindows()
	if len(wins) == 0 {
		return
	}

	const fixedWidth float32 = 280
	pad := theme.Padding() * 2

	text := widget.NewLabel(l.Full)
	text.Wrapping = fyne.TextWrapWord
	// Resize перед MinSize щоб Label розрахував висоту з урахуванням wrapping,
	// інакше MinSize повертає розмір одного символу.
	text.Resize(fyne.NewSize(fixedWidth-pad, 0))

	bg := canvas.NewRectangle(theme.MenuBackgroundColor())
	content := container.NewPadded(text)

	pop := widget.NewPopUp(container.NewStack(bg, content), wins[0].Canvas())
	pop.Resize(fyne.NewSize(fixedWidth, text.MinSize().Height+pad))
	pop.ShowAtPosition(fyne.NewPos(pos.X+12, pos.Y+12))

	l.mu.Lock()
	l.popup = pop
	l.mu.Unlock()
}

func (l *TooltipLabel) hideTooltip() {
	l.mu.Lock()
	pop := l.popup
	l.popup = nil
	l.mu.Unlock()
	if pop != nil {
		pop.Hide()
	}
}

// --- Renderer ---

func (l *TooltipLabel) CreateRenderer() fyne.WidgetRenderer {
	text := canvas.NewText(l.Text, theme.ForegroundColor())
	text.Alignment = l.Align
	text.TextStyle = l.Style
	return &tooltipLabelRenderer{text: text, owner: l}
}

type tooltipLabelRenderer struct {
	text  *canvas.Text
	owner *TooltipLabel
}

func (r *tooltipLabelRenderer) Layout(size fyne.Size) {
	r.text.Move(fyne.NewPos(0, 0))
	r.text.Resize(size)
}

func (r *tooltipLabelRenderer) MinSize() fyne.Size {
	return r.text.MinSize()
}

func (r *tooltipLabelRenderer) Refresh() {
	r.text.Text = r.owner.Text
	r.text.Alignment = r.owner.Align
	r.text.TextStyle = r.owner.Style
	r.text.Color = theme.ForegroundColor()
	canvas.Refresh(r.text)
}

func (r *tooltipLabelRenderer) Destroy() {}

func (r *tooltipLabelRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.text}
}
