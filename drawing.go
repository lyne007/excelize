// Copyright 2016 - 2020 The excelize Authors. All rights reserved. Use of
// this source code is governed by a BSD-style license that can be found in
// the LICENSE file.
//
// Package excelize providing a set of functions that allow you to write to
// and read from XLSX / XLSM / XLTM files. Supports reading and writing
// spreadsheet documents generated by Microsoft Exce™ 2007 and later. Supports
// complex components by high compatibility, and provided streaming API for
// generating or reading data from a worksheet with huge amounts of data. This
// library needs Go version 1.10 or later.

package excelize

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"reflect"
	"strconv"
	"strings"
)

// prepareDrawing provides a function to prepare drawing ID and XML by given
// drawingID, worksheet name and default drawingXML.
func (f *File) prepareDrawing(xlsx *xlsxWorksheet, drawingID int, sheet, drawingXML string) (int, string) {
	sheetRelationshipsDrawingXML := "../drawings/drawing" + strconv.Itoa(drawingID) + ".xml"
	if xlsx.Drawing != nil {
		// The worksheet already has a picture or chart relationships, use the relationships drawing ../drawings/drawing%d.xml.
		sheetRelationshipsDrawingXML = f.getSheetRelationshipsTargetByID(sheet, xlsx.Drawing.RID)
		drawingID, _ = strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(sheetRelationshipsDrawingXML, "../drawings/drawing"), ".xml"))
		drawingXML = strings.Replace(sheetRelationshipsDrawingXML, "..", "xl", -1)
	} else {
		// Add first picture for given sheet.
		sheetRels := "xl/worksheets/_rels/" + strings.TrimPrefix(f.sheetMap[trimSheetName(sheet)], "xl/worksheets/") + ".rels"
		rID := f.addRels(sheetRels, SourceRelationshipDrawingML, sheetRelationshipsDrawingXML, "")
		f.addSheetDrawing(sheet, rID)
	}
	return drawingID, drawingXML
}

// prepareChartSheetDrawing provides a function to prepare drawing ID and XML
// by given drawingID, worksheet name and default drawingXML.
func (f *File) prepareChartSheetDrawing(xlsx *xlsxChartsheet, drawingID int, sheet string) {
	sheetRelationshipsDrawingXML := "../drawings/drawing" + strconv.Itoa(drawingID) + ".xml"
	// Only allow one chart in a chartsheet.
	sheetRels := "xl/chartsheets/_rels/" + strings.TrimPrefix(f.sheetMap[trimSheetName(sheet)], "xl/chartsheets/") + ".rels"
	rID := f.addRels(sheetRels, SourceRelationshipDrawingML, sheetRelationshipsDrawingXML, "")
	xlsx.Drawing = &xlsxDrawing{
		RID: "rId" + strconv.Itoa(rID),
	}
	return
}

// addChart provides a function to create chart as xl/charts/chart%d.xml by
// given format sets.
func (f *File) addChart(formatSet *formatChart, comboCharts []*formatChart) {
	count := f.countCharts()
	xlsxChartSpace := xlsxChartSpace{
		XMLNSc:         NameSpaceDrawingMLChart,
		XMLNSa:         NameSpaceDrawingML,
		XMLNSr:         SourceRelationship,
		XMLNSc16r2:     SourceRelationshipChart201506,
		Date1904:       &attrValBool{Val: boolPtr(false)},
		Lang:           &attrValString{Val: stringPtr("en-US")},
		RoundedCorners: &attrValBool{Val: boolPtr(false)},
		Chart: cChart{
			Title: &cTitle{
				Tx: cTx{
					Rich: &cRich{
						P: aP{
							PPr: &aPPr{
								DefRPr: aRPr{
									Kern:   1200,
									Strike: "noStrike",
									U:      "none",
									Sz:     1400,
									SolidFill: &aSolidFill{
										SchemeClr: &aSchemeClr{
											Val: "tx1",
											LumMod: &attrValInt{
												Val: intPtr(65000),
											},
											LumOff: &attrValInt{
												Val: intPtr(35000),
											},
										},
									},
									Ea: &aEa{
										Typeface: "+mn-ea",
									},
									Cs: &aCs{
										Typeface: "+mn-cs",
									},
									Latin: &aLatin{
										Typeface: "+mn-lt",
									},
								},
							},
							R: &aR{
								RPr: aRPr{
									Lang:    "en-US",
									AltLang: "en-US",
								},
								T: formatSet.Title.Name,
							},
						},
					},
				},
				TxPr: cTxPr{
					P: aP{
						PPr: &aPPr{
							DefRPr: aRPr{
								Kern:   1200,
								U:      "none",
								Sz:     14000,
								Strike: "noStrike",
							},
						},
						EndParaRPr: &aEndParaRPr{
							Lang: "en-US",
						},
					},
				},
				Overlay: &attrValBool{Val: boolPtr(false)},
			},
			View3D: &cView3D{
				RotX:        &attrValInt{Val: intPtr(chartView3DRotX[formatSet.Type])},
				RotY:        &attrValInt{Val: intPtr(chartView3DRotY[formatSet.Type])},
				Perspective: &attrValInt{Val: intPtr(chartView3DPerspective[formatSet.Type])},
				RAngAx:      &attrValInt{Val: intPtr(chartView3DRAngAx[formatSet.Type])},
			},
			Floor: &cThicknessSpPr{
				Thickness: &attrValInt{Val: intPtr(0)},
			},
			SideWall: &cThicknessSpPr{
				Thickness: &attrValInt{Val: intPtr(0)},
			},
			BackWall: &cThicknessSpPr{
				Thickness: &attrValInt{Val: intPtr(0)},
			},
			PlotArea: &cPlotArea{},
			Legend: &cLegend{
				LegendPos: &attrValString{Val: stringPtr(chartLegendPosition[formatSet.Legend.Position])},
				Overlay:   &attrValBool{Val: boolPtr(false)},
			},

			PlotVisOnly:      &attrValBool{Val: boolPtr(false)},
			DispBlanksAs:     &attrValString{Val: stringPtr(formatSet.ShowBlanksAs)},
			ShowDLblsOverMax: &attrValBool{Val: boolPtr(false)},
		},
		SpPr: &cSpPr{
			SolidFill: &aSolidFill{
				SchemeClr: &aSchemeClr{Val: "bg1"},
			},
			Ln: &aLn{
				W:    9525,
				Cap:  "flat",
				Cmpd: "sng",
				Algn: "ctr",
				SolidFill: &aSolidFill{
					SchemeClr: &aSchemeClr{Val: "tx1",
						LumMod: &attrValInt{
							Val: intPtr(15000),
						},
						LumOff: &attrValInt{
							Val: intPtr(85000),
						},
					},
				},
			},
		},
		PrintSettings: &cPrintSettings{
			PageMargins: &cPageMargins{
				B:      0.75,
				L:      0.7,
				R:      0.7,
				T:      0.7,
				Header: 0.3,
				Footer: 0.3,
			},
		},
	}
	plotAreaFunc := map[string]func(*formatChart) *cPlotArea{
		Area:                        f.drawBaseChart,
		AreaStacked:                 f.drawBaseChart,
		AreaPercentStacked:          f.drawBaseChart,
		Area3D:                      f.drawBaseChart,
		Area3DStacked:               f.drawBaseChart,
		Area3DPercentStacked:        f.drawBaseChart,
		Bar:                         f.drawBaseChart,
		BarStacked:                  f.drawBaseChart,
		BarPercentStacked:           f.drawBaseChart,
		Bar3DClustered:              f.drawBaseChart,
		Bar3DStacked:                f.drawBaseChart,
		Bar3DPercentStacked:         f.drawBaseChart,
		Bar3DConeClustered:          f.drawBaseChart,
		Bar3DConeStacked:            f.drawBaseChart,
		Bar3DConePercentStacked:     f.drawBaseChart,
		Bar3DPyramidClustered:       f.drawBaseChart,
		Bar3DPyramidStacked:         f.drawBaseChart,
		Bar3DPyramidPercentStacked:  f.drawBaseChart,
		Bar3DCylinderClustered:      f.drawBaseChart,
		Bar3DCylinderStacked:        f.drawBaseChart,
		Bar3DCylinderPercentStacked: f.drawBaseChart,
		Col:                         f.drawBaseChart,
		ColStacked:                  f.drawBaseChart,
		ColPercentStacked:           f.drawBaseChart,
		Col3D:                       f.drawBaseChart,
		Col3DClustered:              f.drawBaseChart,
		Col3DStacked:                f.drawBaseChart,
		Col3DPercentStacked:         f.drawBaseChart,
		Col3DCone:                   f.drawBaseChart,
		Col3DConeClustered:          f.drawBaseChart,
		Col3DConeStacked:            f.drawBaseChart,
		Col3DConePercentStacked:     f.drawBaseChart,
		Col3DPyramid:                f.drawBaseChart,
		Col3DPyramidClustered:       f.drawBaseChart,
		Col3DPyramidStacked:         f.drawBaseChart,
		Col3DPyramidPercentStacked:  f.drawBaseChart,
		Col3DCylinder:               f.drawBaseChart,
		Col3DCylinderClustered:      f.drawBaseChart,
		Col3DCylinderStacked:        f.drawBaseChart,
		Col3DCylinderPercentStacked: f.drawBaseChart,
		Doughnut:                    f.drawDoughnutChart,
		Line:                        f.drawLineChart,
		Pie3D:                       f.drawPie3DChart,
		Pie:                         f.drawPieChart,
		PieOfPieChart:               f.drawPieOfPieChart,
		BarOfPieChart:               f.drawBarOfPieChart,
		Radar:                       f.drawRadarChart,
		Scatter:                     f.drawScatterChart,
		Surface3D:                   f.drawSurface3DChart,
		WireframeSurface3D:          f.drawSurface3DChart,
		Contour:                     f.drawSurfaceChart,
		WireframeContour:            f.drawSurfaceChart,
		Bubble:                      f.drawBaseChart,
		Bubble3D:                    f.drawBaseChart,
	}
	addChart := func(c, p *cPlotArea) {
		immutable, mutable := reflect.ValueOf(c).Elem(), reflect.ValueOf(p).Elem()
		for i := 0; i < mutable.NumField(); i++ {
			field := mutable.Field(i)
			if field.IsNil() {
				continue
			}
			immutable.FieldByName(mutable.Type().Field(i).Name).Set(field)
		}
	}
	addChart(xlsxChartSpace.Chart.PlotArea, plotAreaFunc[formatSet.Type](formatSet))
	order := len(formatSet.Series)
	for idx := range comboCharts {
		comboCharts[idx].order = order
		addChart(xlsxChartSpace.Chart.PlotArea, plotAreaFunc[comboCharts[idx].Type](comboCharts[idx]))
		order += len(comboCharts[idx].Series)
	}
	chart, _ := xml.Marshal(xlsxChartSpace)
	media := "xl/charts/chart" + strconv.Itoa(count+1) + ".xml"
	f.saveFileList(media, chart)
}

// drawBaseChart provides a function to draw the c:plotArea element for bar,
// and column series charts by given format sets.
func (f *File) drawBaseChart(formatSet *formatChart) *cPlotArea {
	c := cCharts{
		BarDir: &attrValString{
			Val: stringPtr("col"),
		},
		Grouping: &attrValString{
			Val: stringPtr("clustered"),
		},
		VaryColors: &attrValBool{
			Val: boolPtr(true),
		},
		Ser:   f.drawChartSeries(formatSet),
		Shape: f.drawChartShape(formatSet),
		DLbls: f.drawChartDLbls(formatSet),
		AxID: []*attrValInt{
			{Val: intPtr(754001152)},
			{Val: intPtr(753999904)},
		},
		Overlap: &attrValInt{Val: intPtr(100)},
	}
	var ok bool
	if *c.BarDir.Val, ok = plotAreaChartBarDir[formatSet.Type]; !ok {
		c.BarDir = nil
	}
	if *c.Grouping.Val, ok = plotAreaChartGrouping[formatSet.Type]; !ok {
		c.Grouping = nil
	}
	if *c.Overlap.Val, ok = plotAreaChartOverlap[formatSet.Type]; !ok {
		c.Overlap = nil
	}
	catAx := f.drawPlotAreaCatAx(formatSet)
	valAx := f.drawPlotAreaValAx(formatSet)
	charts := map[string]*cPlotArea{
		"area": {
			AreaChart: &c,
			CatAx:     catAx,
			ValAx:     valAx,
		},
		"areaStacked": {
			AreaChart: &c,
			CatAx:     catAx,
			ValAx:     valAx,
		},
		"areaPercentStacked": {
			AreaChart: &c,
			CatAx:     catAx,
			ValAx:     valAx,
		},
		"area3D": {
			Area3DChart: &c,
			CatAx:       catAx,
			ValAx:       valAx,
		},
		"area3DStacked": {
			Area3DChart: &c,
			CatAx:       catAx,
			ValAx:       valAx,
		},
		"area3DPercentStacked": {
			Area3DChart: &c,
			CatAx:       catAx,
			ValAx:       valAx,
		},
		"bar": {
			BarChart: &c,
			CatAx:    catAx,
			ValAx:    valAx,
		},
		"barStacked": {
			BarChart: &c,
			CatAx:    catAx,
			ValAx:    valAx,
		},
		"barPercentStacked": {
			BarChart: &c,
			CatAx:    catAx,
			ValAx:    valAx,
		},
		"bar3DClustered": {
			Bar3DChart: &c,
			CatAx:      catAx,
			ValAx:      valAx,
		},
		"bar3DStacked": {
			Bar3DChart: &c,
			CatAx:      catAx,
			ValAx:      valAx,
		},
		"bar3DPercentStacked": {
			Bar3DChart: &c,
			CatAx:      catAx,
			ValAx:      valAx,
		},
		"bar3DConeClustered": {
			Bar3DChart: &c,
			CatAx:      catAx,
			ValAx:      valAx,
		},
		"bar3DConeStacked": {
			Bar3DChart: &c,
			CatAx:      catAx,
			ValAx:      valAx,
		},
		"bar3DConePercentStacked": {
			Bar3DChart: &c,
			CatAx:      catAx,
			ValAx:      valAx,
		},
		"bar3DPyramidClustered": {
			Bar3DChart: &c,
			CatAx:      catAx,
			ValAx:      valAx,
		},
		"bar3DPyramidStacked": {
			Bar3DChart: &c,
			CatAx:      catAx,
			ValAx:      valAx,
		},
		"bar3DPyramidPercentStacked": {
			Bar3DChart: &c,
			CatAx:      catAx,
			ValAx:      valAx,
		},
		"bar3DCylinderClustered": {
			Bar3DChart: &c,
			CatAx:      catAx,
			ValAx:      valAx,
		},
		"bar3DCylinderStacked": {
			Bar3DChart: &c,
			CatAx:      catAx,
			ValAx:      valAx,
		},
		"bar3DCylinderPercentStacked": {
			Bar3DChart: &c,
			CatAx:      catAx,
			ValAx:      valAx,
		},
		"col": {
			BarChart: &c,
			CatAx:    catAx,
			ValAx:    valAx,
		},
		"colStacked": {
			BarChart: &c,
			CatAx:    catAx,
			ValAx:    valAx,
		},
		"colPercentStacked": {
			BarChart: &c,
			CatAx:    catAx,
			ValAx:    valAx,
		},
		"col3D": {
			Bar3DChart: &c,
			CatAx:      catAx,
			ValAx:      valAx,
		},
		"col3DClustered": {
			Bar3DChart: &c,
			CatAx:      catAx,
			ValAx:      valAx,
		},
		"col3DStacked": {
			Bar3DChart: &c,
			CatAx:      catAx,
			ValAx:      valAx,
		},
		"col3DPercentStacked": {
			Bar3DChart: &c,
			CatAx:      catAx,
			ValAx:      valAx,
		},
		"col3DCone": {
			Bar3DChart: &c,
			CatAx:      catAx,
			ValAx:      valAx,
		},
		"col3DConeClustered": {
			Bar3DChart: &c,
			CatAx:      catAx,
			ValAx:      valAx,
		},
		"col3DConeStacked": {
			Bar3DChart: &c,
			CatAx:      catAx,
			ValAx:      valAx,
		},
		"col3DConePercentStacked": {
			Bar3DChart: &c,
			CatAx:      catAx,
			ValAx:      valAx,
		},
		"col3DPyramid": {
			Bar3DChart: &c,
			CatAx:      catAx,
			ValAx:      valAx,
		},
		"col3DPyramidClustered": {
			Bar3DChart: &c,
			CatAx:      catAx,
			ValAx:      valAx,
		},
		"col3DPyramidStacked": {
			Bar3DChart: &c,
			CatAx:      catAx,
			ValAx:      valAx,
		},
		"col3DPyramidPercentStacked": {
			Bar3DChart: &c,
			CatAx:      catAx,
			ValAx:      valAx,
		},
		"col3DCylinder": {
			Bar3DChart: &c,
			CatAx:      catAx,
			ValAx:      valAx,
		},
		"col3DCylinderClustered": {
			Bar3DChart: &c,
			CatAx:      catAx,
			ValAx:      valAx,
		},
		"col3DCylinderStacked": {
			Bar3DChart: &c,
			CatAx:      catAx,
			ValAx:      valAx,
		},
		"col3DCylinderPercentStacked": {
			Bar3DChart: &c,
			CatAx:      catAx,
			ValAx:      valAx,
		},
		"bubble": {
			BubbleChart: &c,
			CatAx:       catAx,
			ValAx:       valAx,
		},
		"bubble3D": {
			BubbleChart: &c,
			CatAx:       catAx,
			ValAx:       valAx,
		},
	}
	return charts[formatSet.Type]
}

// drawDoughnutChart provides a function to draw the c:plotArea element for
// doughnut chart by given format sets.
func (f *File) drawDoughnutChart(formatSet *formatChart) *cPlotArea {
	return &cPlotArea{
		DoughnutChart: &cCharts{
			VaryColors: &attrValBool{
				Val: boolPtr(true),
			},
			Ser:      f.drawChartSeries(formatSet),
			HoleSize: &attrValInt{Val: intPtr(75)},
		},
	}
}

// drawLineChart provides a function to draw the c:plotArea element for line
// chart by given format sets.
func (f *File) drawLineChart(formatSet *formatChart) *cPlotArea {
	return &cPlotArea{
		LineChart: &cCharts{
			Grouping: &attrValString{
				Val: stringPtr(plotAreaChartGrouping[formatSet.Type]),
			},
			VaryColors: &attrValBool{
				Val: boolPtr(false),
			},
			Ser:   f.drawChartSeries(formatSet),
			DLbls: f.drawChartDLbls(formatSet),
			Smooth: &attrValBool{
				Val: boolPtr(false),
			},
			AxID: []*attrValInt{
				{Val: intPtr(754001152)},
				{Val: intPtr(753999904)},
			},
		},
		CatAx: f.drawPlotAreaCatAx(formatSet),
		ValAx: f.drawPlotAreaValAx(formatSet),
	}
}

// drawPieChart provides a function to draw the c:plotArea element for pie
// chart by given format sets.
func (f *File) drawPieChart(formatSet *formatChart) *cPlotArea {
	return &cPlotArea{
		PieChart: &cCharts{
			VaryColors: &attrValBool{
				Val: boolPtr(true),
			},
			Ser: f.drawChartSeries(formatSet),
		},
	}
}

// drawPie3DChart provides a function to draw the c:plotArea element for 3D
// pie chart by given format sets.
func (f *File) drawPie3DChart(formatSet *formatChart) *cPlotArea {
	return &cPlotArea{
		Pie3DChart: &cCharts{
			VaryColors: &attrValBool{
				Val: boolPtr(true),
			},
			Ser: f.drawChartSeries(formatSet),
		},
	}
}

// drawPieOfPieChart provides a function to draw the c:plotArea element for
// pie chart by given format sets.
func (f *File) drawPieOfPieChart(formatSet *formatChart) *cPlotArea {
	return &cPlotArea{
		OfPieChart: &cCharts{
			OfPieType: &attrValString{
				Val: stringPtr("pie"),
			},
			VaryColors: &attrValBool{
				Val: boolPtr(true),
			},
			Ser:      f.drawChartSeries(formatSet),
			SerLines: &attrValString{},
		},
	}
}

// drawBarOfPieChart provides a function to draw the c:plotArea element for
// pie chart by given format sets.
func (f *File) drawBarOfPieChart(formatSet *formatChart) *cPlotArea {
	return &cPlotArea{
		OfPieChart: &cCharts{
			OfPieType: &attrValString{
				Val: stringPtr("bar"),
			},
			VaryColors: &attrValBool{
				Val: boolPtr(true),
			},
			Ser:      f.drawChartSeries(formatSet),
			SerLines: &attrValString{},
		},
	}
}

// drawRadarChart provides a function to draw the c:plotArea element for radar
// chart by given format sets.
func (f *File) drawRadarChart(formatSet *formatChart) *cPlotArea {
	return &cPlotArea{
		RadarChart: &cCharts{
			RadarStyle: &attrValString{
				Val: stringPtr("marker"),
			},
			VaryColors: &attrValBool{
				Val: boolPtr(false),
			},
			Ser:   f.drawChartSeries(formatSet),
			DLbls: f.drawChartDLbls(formatSet),
			AxID: []*attrValInt{
				{Val: intPtr(754001152)},
				{Val: intPtr(753999904)},
			},
		},
		CatAx: f.drawPlotAreaCatAx(formatSet),
		ValAx: f.drawPlotAreaValAx(formatSet),
	}
}

// drawScatterChart provides a function to draw the c:plotArea element for
// scatter chart by given format sets.
func (f *File) drawScatterChart(formatSet *formatChart) *cPlotArea {
	return &cPlotArea{
		ScatterChart: &cCharts{
			ScatterStyle: &attrValString{
				Val: stringPtr("smoothMarker"), // line,lineMarker,marker,none,smooth,smoothMarker
			},
			VaryColors: &attrValBool{
				Val: boolPtr(false),
			},
			Ser:   f.drawChartSeries(formatSet),
			DLbls: f.drawChartDLbls(formatSet),
			AxID: []*attrValInt{
				{Val: intPtr(754001152)},
				{Val: intPtr(753999904)},
			},
		},
		CatAx: f.drawPlotAreaCatAx(formatSet),
		ValAx: f.drawPlotAreaValAx(formatSet),
	}
}

// drawSurface3DChart provides a function to draw the c:surface3DChart element by
// given format sets.
func (f *File) drawSurface3DChart(formatSet *formatChart) *cPlotArea {
	plotArea := &cPlotArea{
		Surface3DChart: &cCharts{
			Ser: f.drawChartSeries(formatSet),
			AxID: []*attrValInt{
				{Val: intPtr(754001152)},
				{Val: intPtr(753999904)},
				{Val: intPtr(832256642)},
			},
		},
		CatAx: f.drawPlotAreaCatAx(formatSet),
		ValAx: f.drawPlotAreaValAx(formatSet),
		SerAx: f.drawPlotAreaSerAx(formatSet),
	}
	if formatSet.Type == WireframeSurface3D {
		plotArea.Surface3DChart.Wireframe = &attrValBool{Val: boolPtr(true)}
	}
	return plotArea
}

// drawSurfaceChart provides a function to draw the c:surfaceChart element by
// given format sets.
func (f *File) drawSurfaceChart(formatSet *formatChart) *cPlotArea {
	plotArea := &cPlotArea{
		SurfaceChart: &cCharts{
			Ser: f.drawChartSeries(formatSet),
			AxID: []*attrValInt{
				{Val: intPtr(754001152)},
				{Val: intPtr(753999904)},
				{Val: intPtr(832256642)},
			},
		},
		CatAx: f.drawPlotAreaCatAx(formatSet),
		ValAx: f.drawPlotAreaValAx(formatSet),
		SerAx: f.drawPlotAreaSerAx(formatSet),
	}
	if formatSet.Type == WireframeContour {
		plotArea.SurfaceChart.Wireframe = &attrValBool{Val: boolPtr(true)}
	}
	return plotArea
}

// drawChartShape provides a function to draw the c:shape element by given
// format sets.
func (f *File) drawChartShape(formatSet *formatChart) *attrValString {
	shapes := map[string]string{
		Bar3DConeClustered:          "cone",
		Bar3DConeStacked:            "cone",
		Bar3DConePercentStacked:     "cone",
		Bar3DPyramidClustered:       "pyramid",
		Bar3DPyramidStacked:         "pyramid",
		Bar3DPyramidPercentStacked:  "pyramid",
		Bar3DCylinderClustered:      "cylinder",
		Bar3DCylinderStacked:        "cylinder",
		Bar3DCylinderPercentStacked: "cylinder",
		Col3DCone:                   "cone",
		Col3DConeClustered:          "cone",
		Col3DConeStacked:            "cone",
		Col3DConePercentStacked:     "cone",
		Col3DPyramid:                "pyramid",
		Col3DPyramidClustered:       "pyramid",
		Col3DPyramidStacked:         "pyramid",
		Col3DPyramidPercentStacked:  "pyramid",
		Col3DCylinder:               "cylinder",
		Col3DCylinderClustered:      "cylinder",
		Col3DCylinderStacked:        "cylinder",
		Col3DCylinderPercentStacked: "cylinder",
	}
	if shape, ok := shapes[formatSet.Type]; ok {
		return &attrValString{Val: stringPtr(shape)}
	}
	return nil
}

// drawChartSeries provides a function to draw the c:ser element by given
// format sets.
func (f *File) drawChartSeries(formatSet *formatChart) *[]cSer {
	ser := []cSer{}
	for k := range formatSet.Series {
		ser = append(ser, cSer{
			IDx:   &attrValInt{Val: intPtr(k + formatSet.order)},
			Order: &attrValInt{Val: intPtr(k + formatSet.order)},
			Tx: &cTx{
				StrRef: &cStrRef{
					F: formatSet.Series[k].Name,
				},
			},
			SpPr:       f.drawChartSeriesSpPr(k, formatSet),
			Marker:     f.drawChartSeriesMarker(k, formatSet),
			DPt:        f.drawChartSeriesDPt(k, formatSet),
			DLbls:      f.drawChartSeriesDLbls(formatSet),
			Cat:        f.drawChartSeriesCat(formatSet.Series[k], formatSet),
			Val:        f.drawChartSeriesVal(formatSet.Series[k], formatSet),
			XVal:       f.drawChartSeriesXVal(formatSet.Series[k], formatSet),
			YVal:       f.drawChartSeriesYVal(formatSet.Series[k], formatSet),
			BubbleSize: f.drawCharSeriesBubbleSize(formatSet.Series[k], formatSet),
			Bubble3D:   f.drawCharSeriesBubble3D(formatSet),
		})
	}
	return &ser
}

// drawChartSeriesSpPr provides a function to draw the c:spPr element by given
// format sets.
func (f *File) drawChartSeriesSpPr(i int, formatSet *formatChart) *cSpPr {
	spPrScatter := &cSpPr{
		Ln: &aLn{
			W:      25400,
			NoFill: " ",
		},
	}
	spPrLine := &cSpPr{
		Ln: &aLn{
			W:   f.ptToEMUs(formatSet.Series[i].Line.Width),
			Cap: "rnd", // rnd, sq, flat
		},
	}
	if i+formatSet.order < 6 {
		spPrLine.Ln.SolidFill = &aSolidFill{
			SchemeClr: &aSchemeClr{Val: "accent" + strconv.Itoa(i+formatSet.order+1)},
		}
	}
	chartSeriesSpPr := map[string]*cSpPr{Line: spPrLine, Scatter: spPrScatter}
	return chartSeriesSpPr[formatSet.Type]
}

// drawChartSeriesDPt provides a function to draw the c:dPt element by given
// data index and format sets.
func (f *File) drawChartSeriesDPt(i int, formatSet *formatChart) []*cDPt {
	dpt := []*cDPt{{
		IDx:      &attrValInt{Val: intPtr(i)},
		Bubble3D: &attrValBool{Val: boolPtr(false)},
		SpPr: &cSpPr{
			SolidFill: &aSolidFill{
				SchemeClr: &aSchemeClr{Val: "accent" + strconv.Itoa(i+1)},
			},
			Ln: &aLn{
				W:   25400,
				Cap: "rnd",
				SolidFill: &aSolidFill{
					SchemeClr: &aSchemeClr{Val: "lt" + strconv.Itoa(i+1)},
				},
			},
			Sp3D: &aSp3D{
				ContourW: 25400,
				ContourClr: &aContourClr{
					SchemeClr: &aSchemeClr{Val: "lt" + strconv.Itoa(i+1)},
				},
			},
		},
	}}
	chartSeriesDPt := map[string][]*cDPt{Pie: dpt, Pie3D: dpt}
	return chartSeriesDPt[formatSet.Type]
}

// drawChartSeriesCat provides a function to draw the c:cat element by given
// chart series and format sets.
func (f *File) drawChartSeriesCat(v formatChartSeries, formatSet *formatChart) *cCat {
	cat := &cCat{
		StrRef: &cStrRef{
			F: v.Categories,
		},
	}
	chartSeriesCat := map[string]*cCat{Scatter: nil, Bubble: nil, Bubble3D: nil}
	if _, ok := chartSeriesCat[formatSet.Type]; ok || v.Categories == "" {
		return nil
	}
	return cat
}

// drawChartSeriesVal provides a function to draw the c:val element by given
// chart series and format sets.
func (f *File) drawChartSeriesVal(v formatChartSeries, formatSet *formatChart) *cVal {
	val := &cVal{
		NumRef: &cNumRef{
			F: v.Values,
		},
	}
	chartSeriesVal := map[string]*cVal{Scatter: nil, Bubble: nil, Bubble3D: nil}
	if _, ok := chartSeriesVal[formatSet.Type]; ok {
		return nil
	}
	return val
}

// drawChartSeriesMarker provides a function to draw the c:marker element by
// given data index and format sets.
func (f *File) drawChartSeriesMarker(i int, formatSet *formatChart) *cMarker {
	marker := &cMarker{
		Symbol: &attrValString{Val: stringPtr("circle")},
		Size:   &attrValInt{Val: intPtr(5)},
	}
	if i < 6 {
		marker.SpPr = &cSpPr{
			SolidFill: &aSolidFill{
				SchemeClr: &aSchemeClr{
					Val: "accent" + strconv.Itoa(i+1),
				},
			},
			Ln: &aLn{
				W: 9252,
				SolidFill: &aSolidFill{
					SchemeClr: &aSchemeClr{
						Val: "accent" + strconv.Itoa(i+1),
					},
				},
			},
		}
	}
	chartSeriesMarker := map[string]*cMarker{Scatter: marker}
	return chartSeriesMarker[formatSet.Type]
}

// drawChartSeriesXVal provides a function to draw the c:xVal element by given
// chart series and format sets.
func (f *File) drawChartSeriesXVal(v formatChartSeries, formatSet *formatChart) *cCat {
	cat := &cCat{
		StrRef: &cStrRef{
			F: v.Categories,
		},
	}
	chartSeriesXVal := map[string]*cCat{Scatter: cat}
	return chartSeriesXVal[formatSet.Type]
}

// drawChartSeriesYVal provides a function to draw the c:yVal element by given
// chart series and format sets.
func (f *File) drawChartSeriesYVal(v formatChartSeries, formatSet *formatChart) *cVal {
	val := &cVal{
		NumRef: &cNumRef{
			F: v.Values,
		},
	}
	chartSeriesYVal := map[string]*cVal{Scatter: val, Bubble: val, Bubble3D: val}
	return chartSeriesYVal[formatSet.Type]
}

// drawCharSeriesBubbleSize provides a function to draw the c:bubbleSize
// element by given chart series and format sets.
func (f *File) drawCharSeriesBubbleSize(v formatChartSeries, formatSet *formatChart) *cVal {
	if _, ok := map[string]bool{Bubble: true, Bubble3D: true}[formatSet.Type]; !ok {
		return nil
	}
	return &cVal{
		NumRef: &cNumRef{
			F: v.Values,
		},
	}
}

// drawCharSeriesBubble3D provides a function to draw the c:bubble3D element
// by given format sets.
func (f *File) drawCharSeriesBubble3D(formatSet *formatChart) *attrValBool {
	if _, ok := map[string]bool{Bubble3D: true}[formatSet.Type]; !ok {
		return nil
	}
	return &attrValBool{Val: boolPtr(true)}
}

// drawChartDLbls provides a function to draw the c:dLbls element by given
// format sets.
func (f *File) drawChartDLbls(formatSet *formatChart) *cDLbls {
	return &cDLbls{
		ShowLegendKey:   &attrValBool{Val: boolPtr(formatSet.Legend.ShowLegendKey)},
		ShowVal:         &attrValBool{Val: boolPtr(formatSet.Plotarea.ShowVal)},
		ShowCatName:     &attrValBool{Val: boolPtr(formatSet.Plotarea.ShowCatName)},
		ShowSerName:     &attrValBool{Val: boolPtr(formatSet.Plotarea.ShowSerName)},
		ShowBubbleSize:  &attrValBool{Val: boolPtr(formatSet.Plotarea.ShowBubbleSize)},
		ShowPercent:     &attrValBool{Val: boolPtr(formatSet.Plotarea.ShowPercent)},
		ShowLeaderLines: &attrValBool{Val: boolPtr(formatSet.Plotarea.ShowLeaderLines)},
	}
}

// drawChartSeriesDLbls provides a function to draw the c:dLbls element by
// given format sets.
func (f *File) drawChartSeriesDLbls(formatSet *formatChart) *cDLbls {
	dLbls := f.drawChartDLbls(formatSet)
	chartSeriesDLbls := map[string]*cDLbls{Scatter: nil, Surface3D: nil, WireframeSurface3D: nil, Contour: nil, WireframeContour: nil, Bubble: nil, Bubble3D: nil}
	if _, ok := chartSeriesDLbls[formatSet.Type]; ok {
		return nil
	}
	return dLbls
}

// drawPlotAreaCatAx provides a function to draw the c:catAx element.
func (f *File) drawPlotAreaCatAx(formatSet *formatChart) []*cAxs {
	min := &attrValFloat{Val: float64Ptr(formatSet.XAxis.Minimum)}
	max := &attrValFloat{Val: float64Ptr(formatSet.XAxis.Maximum)}
	if formatSet.XAxis.Minimum == 0 {
		min = nil
	}
	if formatSet.XAxis.Maximum == 0 {
		max = nil
	}
	axs := []*cAxs{
		{
			AxID: &attrValInt{Val: intPtr(754001152)},
			Scaling: &cScaling{
				Orientation: &attrValString{Val: stringPtr(orientation[formatSet.XAxis.ReverseOrder])},
				Max:         max,
				Min:         min,
			},
			Delete: &attrValBool{Val: boolPtr(false)},
			AxPos:  &attrValString{Val: stringPtr(catAxPos[formatSet.XAxis.ReverseOrder])},
			NumFmt: &cNumFmt{
				FormatCode:   "General",
				SourceLinked: true,
			},
			MajorTickMark: &attrValString{Val: stringPtr("none")},
			MinorTickMark: &attrValString{Val: stringPtr("none")},
			TickLblPos:    &attrValString{Val: stringPtr("nextTo")},
			SpPr:          f.drawPlotAreaSpPr(),
			TxPr:          f.drawPlotAreaTxPr(),
			CrossAx:       &attrValInt{Val: intPtr(753999904)},
			Crosses:       &attrValString{Val: stringPtr("autoZero")},
			Auto:          &attrValBool{Val: boolPtr(true)},
			LblAlgn:       &attrValString{Val: stringPtr("ctr")},
			LblOffset:     &attrValInt{Val: intPtr(100)},
			NoMultiLvlLbl: &attrValBool{Val: boolPtr(false)},
		},
	}
	if formatSet.XAxis.MajorGridlines {
		axs[0].MajorGridlines = &cChartLines{SpPr: f.drawPlotAreaSpPr()}
	}
	if formatSet.XAxis.MinorGridlines {
		axs[0].MinorGridlines = &cChartLines{SpPr: f.drawPlotAreaSpPr()}
	}
	if formatSet.XAxis.TickLabelSkip != 0 {
		axs[0].TickLblSkip = &attrValInt{Val: intPtr(formatSet.XAxis.TickLabelSkip)}
	}
	return axs
}

// drawPlotAreaValAx provides a function to draw the c:valAx element.
func (f *File) drawPlotAreaValAx(formatSet *formatChart) []*cAxs {
	min := &attrValFloat{Val: float64Ptr(formatSet.YAxis.Minimum)}
	max := &attrValFloat{Val: float64Ptr(formatSet.YAxis.Maximum)}
	if formatSet.YAxis.Minimum == 0 {
		min = nil
	}
	if formatSet.YAxis.Maximum == 0 {
		max = nil
	}
	axs := []*cAxs{
		{
			AxID: &attrValInt{Val: intPtr(753999904)},
			Scaling: &cScaling{
				Orientation: &attrValString{Val: stringPtr(orientation[formatSet.YAxis.ReverseOrder])},
				Max:         max,
				Min:         min,
			},
			Delete: &attrValBool{Val: boolPtr(false)},
			AxPos:  &attrValString{Val: stringPtr(valAxPos[formatSet.YAxis.ReverseOrder])},
			NumFmt: &cNumFmt{
				FormatCode:   chartValAxNumFmtFormatCode[formatSet.Type],
				SourceLinked: true,
			},
			MajorTickMark: &attrValString{Val: stringPtr("none")},
			MinorTickMark: &attrValString{Val: stringPtr("none")},
			TickLblPos:    &attrValString{Val: stringPtr("nextTo")},
			SpPr:          f.drawPlotAreaSpPr(),
			TxPr:          f.drawPlotAreaTxPr(),
			CrossAx:       &attrValInt{Val: intPtr(754001152)},
			Crosses:       &attrValString{Val: stringPtr("autoZero")},
			CrossBetween:  &attrValString{Val: stringPtr(chartValAxCrossBetween[formatSet.Type])},
		},
	}
	if formatSet.YAxis.MajorGridlines {
		axs[0].MajorGridlines = &cChartLines{SpPr: f.drawPlotAreaSpPr()}
	}
	if formatSet.YAxis.MinorGridlines {
		axs[0].MinorGridlines = &cChartLines{SpPr: f.drawPlotAreaSpPr()}
	}
	if pos, ok := valTickLblPos[formatSet.Type]; ok {
		axs[0].TickLblPos.Val = stringPtr(pos)
	}
	if formatSet.YAxis.MajorUnit != 0 {
		axs[0].MajorUnit = &attrValFloat{Val: float64Ptr(formatSet.YAxis.MajorUnit)}
	}
	return axs
}

// drawPlotAreaSerAx provides a function to draw the c:serAx element.
func (f *File) drawPlotAreaSerAx(formatSet *formatChart) []*cAxs {
	min := &attrValFloat{Val: float64Ptr(formatSet.YAxis.Minimum)}
	max := &attrValFloat{Val: float64Ptr(formatSet.YAxis.Maximum)}
	if formatSet.YAxis.Minimum == 0 {
		min = nil
	}
	if formatSet.YAxis.Maximum == 0 {
		max = nil
	}
	return []*cAxs{
		{
			AxID: &attrValInt{Val: intPtr(832256642)},
			Scaling: &cScaling{
				Orientation: &attrValString{Val: stringPtr(orientation[formatSet.YAxis.ReverseOrder])},
				Max:         max,
				Min:         min,
			},
			Delete:     &attrValBool{Val: boolPtr(false)},
			AxPos:      &attrValString{Val: stringPtr(catAxPos[formatSet.XAxis.ReverseOrder])},
			TickLblPos: &attrValString{Val: stringPtr("nextTo")},
			SpPr:       f.drawPlotAreaSpPr(),
			TxPr:       f.drawPlotAreaTxPr(),
			CrossAx:    &attrValInt{Val: intPtr(753999904)},
		},
	}
}

// drawPlotAreaSpPr provides a function to draw the c:spPr element.
func (f *File) drawPlotAreaSpPr() *cSpPr {
	return &cSpPr{
		Ln: &aLn{
			W:    9525,
			Cap:  "flat",
			Cmpd: "sng",
			Algn: "ctr",
			SolidFill: &aSolidFill{
				SchemeClr: &aSchemeClr{
					Val:    "tx1",
					LumMod: &attrValInt{Val: intPtr(15000)},
					LumOff: &attrValInt{Val: intPtr(85000)},
				},
			},
		},
	}
}

// drawPlotAreaTxPr provides a function to draw the c:txPr element.
func (f *File) drawPlotAreaTxPr() *cTxPr {
	return &cTxPr{
		BodyPr: aBodyPr{
			Rot:              -60000000,
			SpcFirstLastPara: true,
			VertOverflow:     "ellipsis",
			Vert:             "horz",
			Wrap:             "square",
			Anchor:           "ctr",
			AnchorCtr:        true,
		},
		P: aP{
			PPr: &aPPr{
				DefRPr: aRPr{
					Sz:       900,
					B:        false,
					I:        false,
					U:        "none",
					Strike:   "noStrike",
					Kern:     1200,
					Baseline: 0,
					SolidFill: &aSolidFill{
						SchemeClr: &aSchemeClr{
							Val:    "tx1",
							LumMod: &attrValInt{Val: intPtr(15000)},
							LumOff: &attrValInt{Val: intPtr(85000)},
						},
					},
					Latin: &aLatin{Typeface: "+mn-lt"},
					Ea:    &aEa{Typeface: "+mn-ea"},
					Cs:    &aCs{Typeface: "+mn-cs"},
				},
			},
			EndParaRPr: &aEndParaRPr{Lang: "en-US"},
		},
	}
}

// drawingParser provides a function to parse drawingXML. In order to solve
// the problem that the label structure is changed after serialization and
// deserialization, two different structures: decodeWsDr and encodeWsDr are
// defined.
func (f *File) drawingParser(path string) (*xlsxWsDr, int) {
	var (
		err error
		ok  bool
	)

	if f.Drawings[path] == nil {
		content := xlsxWsDr{}
		content.A = NameSpaceDrawingML
		content.Xdr = NameSpaceDrawingMLSpreadSheet
		if _, ok = f.XLSX[path]; ok { // Append Model
			decodeWsDr := decodeWsDr{}
			if err = f.xmlNewDecoder(bytes.NewReader(namespaceStrictToTransitional(f.readXML(path)))).
				Decode(&decodeWsDr); err != nil && err != io.EOF {
				log.Printf("xml decode error: %s", err)
			}
			content.R = decodeWsDr.R
			for _, v := range decodeWsDr.OneCellAnchor {
				content.OneCellAnchor = append(content.OneCellAnchor, &xdrCellAnchor{
					EditAs:       v.EditAs,
					GraphicFrame: v.Content,
				})
			}
			for _, v := range decodeWsDr.TwoCellAnchor {
				content.TwoCellAnchor = append(content.TwoCellAnchor, &xdrCellAnchor{
					EditAs:       v.EditAs,
					GraphicFrame: v.Content,
				})
			}
		}
		f.Drawings[path] = &content
	}
	wsDr := f.Drawings[path]
	return wsDr, len(wsDr.OneCellAnchor) + len(wsDr.TwoCellAnchor) + 2
}

// addDrawingChart provides a function to add chart graphic frame by given
// sheet, drawingXML, cell, width, height, relationship index and format sets.
func (f *File) addDrawingChart(sheet, drawingXML, cell string, width, height, rID int, formatSet *formatPicture) error {
	col, row, err := CellNameToCoordinates(cell)
	if err != nil {
		return err
	}
	colIdx := col - 1
	rowIdx := row - 1

	width = int(float64(width) * formatSet.XScale)
	height = int(float64(height) * formatSet.YScale)
	colStart, rowStart, _, _, colEnd, rowEnd, x2, y2 :=
		f.positionObjectPixels(sheet, colIdx, rowIdx, formatSet.OffsetX, formatSet.OffsetY, width, height)
	content, cNvPrID := f.drawingParser(drawingXML)
	twoCellAnchor := xdrCellAnchor{}
	twoCellAnchor.EditAs = formatSet.Positioning
	from := xlsxFrom{}
	from.Col = colStart
	from.ColOff = formatSet.OffsetX * EMU
	from.Row = rowStart
	from.RowOff = formatSet.OffsetY * EMU
	to := xlsxTo{}
	to.Col = colEnd
	to.ColOff = x2 * EMU
	to.Row = rowEnd
	to.RowOff = y2 * EMU
	twoCellAnchor.From = &from
	twoCellAnchor.To = &to

	graphicFrame := xlsxGraphicFrame{
		NvGraphicFramePr: xlsxNvGraphicFramePr{
			CNvPr: &xlsxCNvPr{
				ID:   cNvPrID,
				Name: "Chart " + strconv.Itoa(cNvPrID),
			},
		},
		Graphic: &xlsxGraphic{
			GraphicData: &xlsxGraphicData{
				URI: NameSpaceDrawingMLChart,
				Chart: &xlsxChart{
					C:   NameSpaceDrawingMLChart,
					R:   SourceRelationship,
					RID: "rId" + strconv.Itoa(rID),
				},
			},
		},
	}
	graphic, _ := xml.Marshal(graphicFrame)
	twoCellAnchor.GraphicFrame = string(graphic)
	twoCellAnchor.ClientData = &xdrClientData{
		FLocksWithSheet:  formatSet.FLocksWithSheet,
		FPrintsWithSheet: formatSet.FPrintsWithSheet,
	}
	content.TwoCellAnchor = append(content.TwoCellAnchor, &twoCellAnchor)
	f.Drawings[drawingXML] = content
	return err
}

// addSheetDrawingChart provides a function to add chart graphic frame for
// chartsheet by given sheet, drawingXML, width, height, relationship index
// and format sets.
func (f *File) addSheetDrawingChart(drawingXML string, rID int, formatSet *formatPicture) {
	content, cNvPrID := f.drawingParser(drawingXML)
	absoluteAnchor := xdrCellAnchor{
		EditAs: formatSet.Positioning,
		Pos:    &xlsxPoint2D{},
		Ext:    &xlsxExt{},
	}

	graphicFrame := xlsxGraphicFrame{
		NvGraphicFramePr: xlsxNvGraphicFramePr{
			CNvPr: &xlsxCNvPr{
				ID:   cNvPrID,
				Name: "Chart " + strconv.Itoa(cNvPrID),
			},
		},
		Graphic: &xlsxGraphic{
			GraphicData: &xlsxGraphicData{
				URI: NameSpaceDrawingMLChart,
				Chart: &xlsxChart{
					C:   NameSpaceDrawingMLChart,
					R:   SourceRelationship,
					RID: "rId" + strconv.Itoa(rID),
				},
			},
		},
	}
	graphic, _ := xml.Marshal(graphicFrame)
	absoluteAnchor.GraphicFrame = string(graphic)
	absoluteAnchor.ClientData = &xdrClientData{
		FLocksWithSheet:  formatSet.FLocksWithSheet,
		FPrintsWithSheet: formatSet.FPrintsWithSheet,
	}
	content.AbsoluteAnchor = append(content.AbsoluteAnchor, &absoluteAnchor)
	f.Drawings[drawingXML] = content
	return
}

// deleteDrawing provides a function to delete chart graphic frame by given by
// given coordinates and graphic type.
func (f *File) deleteDrawing(col, row int, drawingXML, drawingType string) (err error) {
	var (
		wsDr            *xlsxWsDr
		deTwoCellAnchor *decodeTwoCellAnchor
	)
	xdrCellAnchorFuncs := map[string]func(anchor *xdrCellAnchor) bool{
		"Chart": func(anchor *xdrCellAnchor) bool { return anchor.Pic == nil },
		"Pic":   func(anchor *xdrCellAnchor) bool { return anchor.Pic != nil },
	}
	decodeTwoCellAnchorFuncs := map[string]func(anchor *decodeTwoCellAnchor) bool{
		"Chart": func(anchor *decodeTwoCellAnchor) bool { return anchor.Pic == nil },
		"Pic":   func(anchor *decodeTwoCellAnchor) bool { return anchor.Pic != nil },
	}
	wsDr, _ = f.drawingParser(drawingXML)
	for idx := 0; idx < len(wsDr.TwoCellAnchor); idx++ {
		if err = nil; wsDr.TwoCellAnchor[idx].From != nil && xdrCellAnchorFuncs[drawingType](wsDr.TwoCellAnchor[idx]) {
			if wsDr.TwoCellAnchor[idx].From.Col == col && wsDr.TwoCellAnchor[idx].From.Row == row {
				wsDr.TwoCellAnchor = append(wsDr.TwoCellAnchor[:idx], wsDr.TwoCellAnchor[idx+1:]...)
				idx--
			}
		}
	}
	for idx := 0; idx < len(wsDr.TwoCellAnchor); idx++ {
		deTwoCellAnchor = new(decodeTwoCellAnchor)
		if err = f.xmlNewDecoder(strings.NewReader("<decodeTwoCellAnchor>" + wsDr.TwoCellAnchor[idx].GraphicFrame + "</decodeTwoCellAnchor>")).
			Decode(deTwoCellAnchor); err != nil && err != io.EOF {
			err = fmt.Errorf("xml decode error: %s", err)
			return
		}
		if err = nil; deTwoCellAnchor.From != nil && decodeTwoCellAnchorFuncs[drawingType](deTwoCellAnchor) {
			if deTwoCellAnchor.From.Col == col && deTwoCellAnchor.From.Row == row {
				wsDr.TwoCellAnchor = append(wsDr.TwoCellAnchor[:idx], wsDr.TwoCellAnchor[idx+1:]...)
				idx--
			}
		}
	}
	f.Drawings[drawingXML] = wsDr
	return err
}
