package branddir

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"math"
	"slices"
	"strconv"
	"strings"
)

// fonts is the fixed font set, in the order of web/packages/tokens/fonts.json, which gives the font files.
var fonts = []string{
	"Inter Variable", "JetBrains Mono Variable", "Figtree Variable", "Public Sans Variable", "Geist Variable",
	"DM Sans Variable", "Bricolage Grotesque Variable", "Source Serif 4 Variable", "Unbounded Variable",
}

// genericFamilies are the CSS generic font families, which a font token can give after its font.
var genericFamilies = []string{
	"serif", "sans-serif", "monospace", "cursive", "fantasy", "system-ui", "ui-serif", "ui-sans-serif",
	"ui-monospace", "ui-rounded", "math", "emoji", "fangsong",
}

type tokenGroup struct {
	name   string
	dtcg   string
	tokens []string
	toCSS  func(json.RawMessage) (string, error)
}

// contract has the groups and tokens of a brand of the design system (web/packages/tokens/tokens/brand/lab.json), in
// the order of the stylesheet.
var contract = []tokenGroup{
	{
		name: "color", dtcg: "color", toCSS: colorCSS,
		tokens: []string{
			"page", "surface", "raised", "backdrop", "fg", "fg-muted", "fg-on-primary", "heading", "link", "brand",
			"fg-on-brand", "hero", "fg-on-hero", "accent", "border", "border-strong",
			"focus", "primary", "primary-hover", "primary-pressed", "secondary",
			"success-bg", "success-border", "success-fg", "warning-bg", "warning-border", "warning-fg",
			"danger-bg", "danger-border", "danger-fg", "info-bg", "info-border", "info-fg",
		},
	},
	{name: "font", dtcg: "fontFamily", toCSS: fontCSS, tokens: []string{"body", "code", "heading"}},
	{name: "radius", dtcg: "dimension", toCSS: dimensionCSS, tokens: []string{"sm", "md", "lg", "full"}},
}

// consoleProperties are the only roles of the brand in the console: the roles of the console shell, the link and the
// primary action (web/packages/ui/design/tokens-and-brands.md, UI-0035). The console uses the lab brand for the rest.
var consoleProperties = []string{
	"--ds-color-brand", "--ds-color-fg-on-brand", "--ds-color-hero", "--ds-color-fg-on-hero", "--ds-color-accent",
	"--ds-color-heading", "--ds-color-link", "--ds-color-primary", "--ds-color-primary-hover",
	"--ds-color-primary-pressed", "--ds-color-fg-on-primary",
}

// parseTokens gives the stylesheets of the brand: the custom properties of the design system with the values of the
// brand, and the properties of the console only. Each value comes only from numbers and fixed names, so a brand
// directory cannot add other CSS.
func parseTokens(data []byte) (styles, consoleStyles []byte, errs []error) {
	var groups map[string]json.RawMessage
	if err := decodeStrict(data, &groups); err != nil {
		return nil, nil, []error{err}
	}

	for _, name := range slices.Sorted(maps.Keys(groups)) {
		if !slices.ContainsFunc(contract, func(g tokenGroup) bool { return g.name == name }) {
			errs = append(errs, fmt.Errorf("%s: the contract has no group of this name", name))
		}
	}

	var css, console strings.Builder

	css.WriteString(":root:root {\n")
	console.WriteString(":root:root {\n")

	for _, group := range contract {
		raw, found := groups[group.name]
		if !found {
			errs = append(errs, fmt.Errorf("%s: the group is missing", group.name))
			continue
		}

		values, groupErrs := parseGroup(group, raw)
		errs = append(errs, groupErrs...)

		for _, token := range group.tokens {
			if value, found := values[token]; found {
				declaration := fmt.Sprintf("  --ds-%s-%s: %s;\n", group.name, token, value)
				css.WriteString(declaration)

				if slices.Contains(consoleProperties, "--ds-"+group.name+"-"+token) {
					console.WriteString(declaration)
				}
			}
		}
	}

	css.WriteString("}\n")
	console.WriteString("}\n")

	return []byte(css.String()), []byte(console.String()), errs
}

// parseGroup gives the CSS value of each valid token of the group.
func parseGroup(group tokenGroup, raw json.RawMessage) (map[string]string, []error) {
	var members map[string]json.RawMessage
	if err := json.Unmarshal(raw, &members); err != nil {
		return nil, []error{fmt.Errorf("%s: the group is not a JSON object", group.name)}
	}

	var errs []error

	var dtcgType string
	if err := json.Unmarshal(members["$type"], &dtcgType); err != nil || dtcgType != group.dtcg {
		errs = append(errs, fmt.Errorf("%s: $type is not %q", group.name, group.dtcg))
	}

	if err := checkDescription(members); err != nil {
		errs = append(errs, fmt.Errorf("%s: %w", group.name, err))
	}

	values := map[string]string{}

	for _, name := range slices.Sorted(maps.Keys(members)) {
		path := group.name + "." + name

		switch {
		case name == "$type" || name == "$description":
		case !slices.Contains(group.tokens, name):
			errs = append(errs, fmt.Errorf("%s: the contract has no token of this name", path))
		default:
			value, err := parseToken(group, members[name])
			if err != nil {
				errs = append(errs, fmt.Errorf("%s: %w", path, err))
				continue
			}

			values[name] = value
		}
	}

	for _, name := range group.tokens {
		if _, found := members[name]; !found {
			errs = append(errs, fmt.Errorf("%s.%s: the token is missing", group.name, name))
		}
	}

	return values, errs
}

func checkDescription(members map[string]json.RawMessage) error {
	raw, found := members["$description"]
	if !found {
		return nil
	}

	var description string
	if err := json.Unmarshal(raw, &description); err != nil {
		return errors.New("$description is not a string")
	}

	return nil
}

type token struct {
	Value       json.RawMessage `json:"$value"`
	Description *string         `json:"$description"`
}

func parseToken(group tokenGroup, raw json.RawMessage) (string, error) {
	var t token
	if err := decodeStrict(raw, &t); err != nil {
		return "", fmt.Errorf("the token has only $value and an optional $description: %w", err)
	}

	if t.Value == nil {
		return "", errors.New("the token has no $value")
	}

	return group.toCSS(t.Value)
}

type colorValue struct {
	ColorSpace string    `json:"colorSpace"`
	Components []float64 `json:"components"`
	Alpha      *float64  `json:"alpha"`
}

func colorCSS(raw json.RawMessage) (string, error) {
	var c colorValue
	if err := decodeStrict(raw, &c); err != nil {
		return "", errors.New("$value is not a literal OKLCH colour: give colorSpace, components and an optional alpha")
	}

	if c.ColorSpace != "oklch" {
		return "", errors.New("the colour space is not oklch")
	}

	if len(c.Components) != 3 {
		return "", errors.New("give 3 components: lightness, chroma and hue")
	}

	l, ch, h := c.Components[0], c.Components[1], c.Components[2]

	switch {
	case l < 0 || l > 1:
		return "", errors.New("the lightness is not from 0 to 1")
	case ch < 0:
		return "", errors.New("the chroma is less than 0")
	case h < 0 || h >= 360:
		return "", errors.New("the hue is not from 0 to less than 360")
	case c.Alpha != nil && (*c.Alpha < 0 || *c.Alpha > 1):
		return "", errors.New("the alpha is not from 0 to 1")
	case !inSRGBGamut(l, ch, h):
		return "", errors.New("the colour is not in the sRGB gamut")
	}

	value := "oklch(" + number(l) + " " + number(ch) + " " + number(h)
	if c.Alpha != nil && *c.Alpha != 1 {
		value += " / " + number(*c.Alpha)
	}

	return value + ")", nil
}

// gamutEpsilon is the tolerance of the inGamut function of colorjs.io, which the max-gamut lint of the design system
// uses. Thus a colour that the lint accepts is also valid here.
const gamutEpsilon = 0.000075

func inSRGBGamut(lightness, chroma, hue float64) bool {
	for _, v := range oklchToSRGB(lightness, chroma, hue) {
		if v < -gamutEpsilon || v > 1+gamutEpsilon {
			return false
		}
	}

	return true
}

// oklchToSRGB converts the colour to gamma-encoded sRGB with the OKLab matrices of CSS Color 4. A channel of a colour
// outside the gamut is less than 0 or more than 1.
func oklchToSRGB(lightness, chroma, hue float64) [3]float64 {
	a := chroma * math.Cos(hue*math.Pi/180)
	b := chroma * math.Sin(hue*math.Pi/180)

	l := cube(lightness + 0.3963377773761749*a + 0.2158037573099136*b)
	m := cube(lightness - 0.1055613458156586*a - 0.0638541728258133*b)
	s := cube(lightness - 0.0894841775298119*a - 1.2914855480194092*b)

	return [3]float64{
		gammaEncode(4.0767416360759583*l - 3.3077115392580629*m + 0.2309699031821043*s),
		gammaEncode(-1.2684379732850315*l + 2.6097573492876882*m - 0.3413193760026570*s),
		gammaEncode(-0.0041960761386756*l - 0.7034186179359362*m + 1.7076146940746117*s),
	}
}

// hexColours gives each opaque colour of a valid tokens file in hex notation, for an email client that does not
// support OKLCH. Each channel is clamped, because the gamut check permits a small tolerance.
func hexColours(data []byte) map[string]string {
	var tokens struct {
		Color map[string]json.RawMessage `json:"color"`
	}
	if err := json.Unmarshal(data, &tokens); err != nil {
		return nil
	}

	colours := map[string]string{}

	for name, member := range tokens.Color {
		var t token
		if strings.HasPrefix(name, "$") || json.Unmarshal(member, &t) != nil {
			continue
		}

		var c colorValue
		if json.Unmarshal(t.Value, &c) != nil || len(c.Components) != 3 || (c.Alpha != nil && *c.Alpha != 1) {
			continue
		}

		rgb := oklchToSRGB(c.Components[0], c.Components[1], c.Components[2])
		colours[name] = fmt.Sprintf("#%02x%02x%02x", channel(rgb[0]), channel(rgb[1]), channel(rgb[2]))
	}

	return colours
}

func channel(v float64) int { return int(math.Round(255 * min(max(v, 0), 1))) }

func cube(v float64) float64 { return v * v * v }

// gammaEncode is the sRGB transfer function (CSS Color 4), also for a negative value.
func gammaEncode(linear float64) float64 {
	magnitude := math.Abs(linear)
	if magnitude <= 0.0031308 {
		return linear * 12.92
	}

	return math.Copysign(1.055*math.Pow(magnitude, 1/2.4)-0.055, linear)
}

func fontCSS(raw json.RawMessage) (string, error) {
	var families []string

	var one string
	if err := json.Unmarshal(raw, &one); err == nil {
		families = []string{one}
	} else if err := json.Unmarshal(raw, &families); err != nil {
		return "", errors.New("$value is not a font family: give a string or a list of strings")
	}

	if len(families) == 0 {
		return "", errors.New("give one font or more")
	}

	if !slices.Contains(fonts, families[0]) {
		return "", fmt.Errorf("%q is not a font of the data holder: give one of %s", families[0], quoted(fonts))
	}

	for _, family := range families[1:] {
		if !slices.Contains(genericFamilies, family) {
			return "", fmt.Errorf("%q is not a generic font family: give only generic families after the font", family)
		}
	}

	return strings.Join(append([]string{strconv.Quote(families[0])}, families[1:]...), ", "), nil
}

func quoted(names []string) string {
	q := make([]string, len(names))
	for i, name := range names {
		q[i] = strconv.Quote(name)
	}

	return strings.Join(q, ", ")
}

type dimensionValue struct {
	Value *float64 `json:"value"`
	Unit  string   `json:"unit"`
}

func dimensionCSS(raw json.RawMessage) (string, error) {
	var d dimensionValue
	if err := decodeStrict(raw, &d); err != nil || d.Value == nil {
		return "", errors.New("$value is not a dimension: give value and unit")
	}

	if d.Unit != "px" && d.Unit != "rem" {
		return "", errors.New(`the unit is not "px" or "rem"`)
	}

	if *d.Value < 0 {
		return "", errors.New("the value is less than 0")
	}

	return number(*d.Value) + d.Unit, nil
}

func number(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}
