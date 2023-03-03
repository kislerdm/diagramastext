/**
 * Function to convert C4 diagram as text to the request
 * string to generate the diagram as png, or svg.
 *
 * Example: the diagram
 * `@startuml
 * Bob -> Alice : hi
 * @enduml`
 *
 * will be converted to 'SoWkIImgAStDuNBAJrBGjLDmpCbCJbMmKiZ8v798pKi1oW00'
 *
 * The resulting string to be used to generate C4 diagram
 * - as png: GET www.plantuml.com/plantuml/png/SoWkIImgAStDuNBAJrBGjLDmpCbCJbMmKiZ8v798pKi1oW00
 * - as svg: GET www.plantuml.com/plantuml/svg/SoWkIImgAStDuNBAJrBGjLDmpCbCJbMmKiZ8v798pKi1oW00
 *
 * @param s C4 Diagram as test.
 * @returns (String)
 */

function compress(s) {
    s = unescape(encodeURIComponent(s));
    var arr = [];
    for (var i = 0; i < s.length; i++)
        arr.push(s.charCodeAt(i));
    var compressor = new Zopfli.RawDeflate(arr);
    var compressed = compressor.compress();
    return encode64_(compressed);
}
