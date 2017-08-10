<?xml version="1.0" encoding="UTF-8"?>
<!-- Generated LXT parser v3.0: do not alter directly -->

<xsl:stylesheet version="1.0" xmlns:install="http://www.microsoft.com/support" xmlns:msxsl="urn:schemas-microsoft-com:xslt" xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsl="http://www.w3.org/1999/XSL/Transform">

  <xsl:variable name="nl">
    <xsl:text>
</xsl:text>
  </xsl:variable>


  <xsl:output encoding="UTF-8" indent="yes" media-type="application/xml" method="text" version="1.0" />

  <xsl:template match="/">
    <xsl:text>// mostly computer generated from from http://standards.iso.org/ittf/PubliclyAvailableStandards/MPEG-DASH_schema_files/DASH-MPD.xsd</xsl:text>
    <xsl:value-of select="$nl" />
    <xsl:value-of select="$nl" />
    <xsl:apply-templates />
  </xsl:template>

  <xsl:template match="/xs:schema">
    <xsl:for-each select="xs:simpleType">
      <xsl:call-template name="simpleType" />
    </xsl:for-each>
    <xsl:for-each select="xs:complexType">
      <xsl:call-template name="complexType" />
    </xsl:for-each>
  </xsl:template>

  <xsl:template name="stripType">
    <xsl:param name="type" select="@type" />

    <xsl:choose>
      <xsl:when test="substring($type,string-length($type) - 3) = &quot;type&quot;">
        <xsl:value-of select="substring($type,1,string-length($type) - 4)" />
      </xsl:when>
      <xsl:when test="substring($type,string-length($type) - 3) = &quot;Type&quot;">
        <xsl:value-of select="substring($type,1,string-length($type) - 4)" />
      </xsl:when>
      <xsl:otherwise>
        <xsl:value-of select="$type" />
      </xsl:otherwise>
    </xsl:choose>
  </xsl:template>

  <xsl:template name="mapType">
    <xsl:param name="type" select="@type" />

    <xsl:choose>
      <xsl:when test="$type = &quot;xs:boolean&quot;">bool</xsl:when>
      <xsl:when test="$type = &quot;xs:string&quot;">string</xsl:when>
      <xsl:when test="$type = &quot;xs:anyURI&quot;">string</xsl:when>
      <xsl:when test="$type = &quot;xs:language&quot;">string</xsl:when>
      <xsl:when test="$type = &quot;xs:integer&quot;">int</xsl:when>
      <xsl:when test="$type = &quot;xs:unsignedInt&quot;">uint</xsl:when>
      <xsl:when test="$type = &quot;xs:unsignedLong&quot;">uint64</xsl:when>
      <xsl:when test="$type = &quot;xs:double&quot;">float64</xsl:when>
      <xsl:when test="$type = &quot;xs:dateTime&quot;">time.Time</xsl:when>
      <xsl:when test="$type = &quot;xs:duration&quot;">Duration</xsl:when>
      <xsl:when test="count(//xs:complexType[@name = $type])">
        <xsl:text>*</xsl:text>
        <xsl:call-template name="stripType">
          <xsl:with-param name="type" select="$type" />
        </xsl:call-template>
      </xsl:when>
      <xsl:otherwise>
        <xsl:call-template name="stripType">
          <xsl:with-param name="type" select="$type" />
        </xsl:call-template>
      </xsl:otherwise>
    </xsl:choose>
  </xsl:template>

  <xsl:template name="unionType">
    <xsl:param name="members"></xsl:param>

    <xsl:variable name="before" select="substring-before($members,&quot; &quot;)" />
    <xsl:variable name="type">
      <xsl:choose>
        <xsl:when test="$before = &quot;&quot;">
          <xsl:value-of select="$members" />
        </xsl:when>
        <xsl:otherwise>
          <xsl:value-of select="$before" />
        </xsl:otherwise>
      </xsl:choose>
    </xsl:variable>
    <xsl:variable name="after" select="substring-after($members,&quot; &quot;)" />
    <xsl:text>	// </xsl:text>
    <xsl:call-template name="mapType">
      <xsl:with-param name="type" select="$type" />
    </xsl:call-template>
    <xsl:value-of select="$nl" />
    <xsl:if test="$after != &quot;&quot;">
      <xsl:call-template name="unionType">
        <xsl:with-param name="members" select="$after" />
      </xsl:call-template>
    </xsl:if>
  </xsl:template>

  <xsl:template name="simpleType">
    <xsl:variable name="name">
      <xsl:call-template name="stripType">
        <xsl:with-param name="type" select="@name" />
      </xsl:call-template>
    </xsl:variable>
    <xsl:text>type </xsl:text>
    <xsl:value-of select="$name" />
    <xsl:text> </xsl:text>
    <xsl:choose>
      <xsl:when test="count(xs:list) &gt; 0">
        <xsl:text>[]</xsl:text>
        <xsl:call-template name="mapType">
          <xsl:with-param name="type" select="xs:list/@itemType" />
        </xsl:call-template>
      </xsl:when>
      <xsl:when test="count(xs:union) &gt; 0">
        <xsl:text>string // union {</xsl:text>
        <xsl:value-of select="$nl" />
        <xsl:call-template name="unionType">
          <xsl:with-param name="members" select="xs:union/@memberTypes" />
        </xsl:call-template>
        <xsl:text>// }</xsl:text>
      </xsl:when>
      <xsl:otherwise>
        <xsl:call-template name="mapType">
          <xsl:with-param name="type" select="xs:restriction/@base" />
        </xsl:call-template>
      </xsl:otherwise>
    </xsl:choose>
    <xsl:value-of select="$nl" />
    <xsl:for-each select="xs:restriction">
      <xsl:call-template name="restrictions">
        <xsl:with-param name="name" select="$name" />
      </xsl:call-template>
    </xsl:for-each>
    <xsl:value-of select="$nl" />
  </xsl:template>

  <xsl:template name="restrictions">
    <xsl:param name="name" select="@name" />

    <xsl:if test="count(xs:enumeration) &gt; 0">
      <xsl:text>var </xsl:text>
      <xsl:value-of select="$name" />
      <xsl:text>_Valid = map[string]bool {</xsl:text>
      <xsl:value-of select="$nl" />
      <xsl:for-each select="xs:enumeration">
        <xsl:text>	&quot;</xsl:text>
        <xsl:value-of select="@value" />
        <xsl:text>&quot;: true,</xsl:text>
        <xsl:value-of select="$nl" />
      </xsl:for-each>
      <xsl:text>}</xsl:text>
    </xsl:if>
    <xsl:if test="count(xs:pattern) &gt; 0">
      <xsl:text>var </xsl:text>
      <xsl:value-of select="$name" />
      <xsl:text>_Validate = regexp.MustCompile(`</xsl:text>
      <xsl:value-of select="xs:pattern/@value" />
      <xsl:text>`)</xsl:text>
    </xsl:if>
    <xsl:if test="count(xs:minInclusive) + count(xs:maxInclusive) + count(xs:minExclusive) + count(xs:maxExclusive)">
      <xsl:text>const (</xsl:text>
      <xsl:value-of select="$nl" />
      <xsl:if test="count(xs:minInclusive)">
        <xsl:text>	</xsl:text>
        <xsl:value-of select="$name" />
        <xsl:text>_MinInclusive </xsl:text>
        <xsl:value-of select="$name" />
        <xsl:text> = </xsl:text>
        <xsl:value-of select="xs:minInclusive/@value" />
        <xsl:value-of select="$nl" />
      </xsl:if>
      <xsl:if test="count(xs:maxInclusive)">
        <xsl:text>	</xsl:text>
        <xsl:value-of select="$name" />
        <xsl:text>_MaxInclusive </xsl:text>
        <xsl:value-of select="$name" />
        <xsl:text> = </xsl:text>
        <xsl:value-of select="xs:maxInclusive/@value" />
        <xsl:value-of select="$nl" />
      </xsl:if>
      <xsl:if test="count(xs:minExclusive)">
        <xsl:text>	</xsl:text>
        <xsl:value-of select="$name" />
        <xsl:text>_MinExclusive </xsl:text>
        <xsl:value-of select="$name" />
        <xsl:text> = </xsl:text>
        <xsl:value-of select="xs:minExclusive/@value" />
        <xsl:value-of select="$nl" />
      </xsl:if>
      <xsl:if test="count(xs:maxExclusive)">
        <xsl:text>	</xsl:text>
        <xsl:value-of select="$name" />
        <xsl:text>_MaxExclusive </xsl:text>
        <xsl:value-of select="$name" />
        <xsl:text> = </xsl:text>
        <xsl:value-of select="xs:maxExclusive/@value" />
        <xsl:value-of select="$nl" />
      </xsl:if>
      <xsl:text>)</xsl:text>
    </xsl:if>
    <xsl:value-of select="$nl" />
  </xsl:template>

  <xsl:template name="exportName">
    <xsl:param name="name" select="@name" />

    <xsl:value-of select="concat(translate(substring($name,1,1),&quot;abcdefghijklmnopqrstuvwxyz&quot;,&quot;ABCDEFGHIJKLMNOPQRSTUVWYYZ&quot;),substring($name,2))" />
  </xsl:template>

  <xsl:template name="simpleContent">
    <xsl:for-each select="xs:attribute[@name]">
      <xsl:text>	</xsl:text>
      <xsl:call-template name="exportName">
        <xsl:with-param name="name" select="@name" />
      </xsl:call-template>
      <xsl:text> </xsl:text>
      <xsl:call-template name="mapType">
        <xsl:with-param name="type" select="@type" />
      </xsl:call-template>
      <xsl:text> `xml:&quot;</xsl:text>
      <xsl:value-of select="@name" />
      <xsl:text>,attr</xsl:text>
      <xsl:if test="count(@use) = 0 or @use != &quot;required&quot;">
        <xsl:text>,omitempty</xsl:text>
      </xsl:if>
      <xsl:text>&quot;`</xsl:text>
      <xsl:if test="count(@default)">
        <xsl:text>	// default: </xsl:text>
        <xsl:value-of select="@default" />
      </xsl:if>
      <xsl:value-of select="$nl" />
    </xsl:for-each>
  </xsl:template>

  <xsl:template name="complexContent">
    <xsl:for-each select="xs:sequence/xs:element">
      <xsl:text>	</xsl:text>
      <xsl:call-template name="exportName">
        <xsl:with-param name="name" select="@name" />
      </xsl:call-template>
      <xsl:text> </xsl:text>
      <xsl:if test="@maxOccurs = &quot;unbounded&quot; or @maxOccurs &gt; 1">
        <xsl:text>[]</xsl:text>
      </xsl:if>
      <xsl:choose>
        <xsl:when test="count(xs:complexType) &gt; 0">
          <xsl:text>struct {</xsl:text>
          <xsl:value-of select="$nl" />
          <xsl:for-each select="xs:complexType">
            <xsl:call-template name="complexContent" />
          </xsl:for-each>
          <xsl:text>	}</xsl:text>
          <xsl:value-of select="$nl" />
        </xsl:when>
        <xsl:otherwise>
          <xsl:call-template name="mapType">
            <xsl:with-param name="type" select="@type" />
          </xsl:call-template>
          <xsl:text> `xml:&quot;</xsl:text>
          <xsl:value-of select="@name" />
          <xsl:if test="count(@minOccurs) and @minOccurs = 0">,omitempty</xsl:if>
          <xsl:text>&quot;`</xsl:text>
          <xsl:value-of select="$nl" />
        </xsl:otherwise>
      </xsl:choose>
    </xsl:for-each>
    <xsl:if test="count(xs:sequence/xs:element) &gt; 0 and count(xs:attribute[@name]) &gt; 0">
      <xsl:value-of select="$nl" />
    </xsl:if>
    <xsl:call-template name="simpleContent" />
  </xsl:template>

  <xsl:template name="complexType">
    <xsl:text>type </xsl:text>
    <xsl:call-template name="stripType">
      <xsl:with-param name="type" select="@name" />
    </xsl:call-template>
    <xsl:text> struct {</xsl:text>
    <xsl:value-of select="$nl" />
    <xsl:for-each select="xs:complexContent/xs:extension">
      <xsl:text>	</xsl:text>
      <xsl:call-template name="mapType">
        <xsl:with-param name="type" select="@base" />
      </xsl:call-template>
      <xsl:value-of select="$nl" />
      <xsl:value-of select="$nl" />
      <xsl:call-template name="complexContent" />
    </xsl:for-each>
    <xsl:for-each select="xs:simpleContent/xs:extension">
      <xsl:text>	CDATA </xsl:text>
      <xsl:call-template name="mapType">
        <xsl:with-param name="type" select="@base" />
      </xsl:call-template>
      <xsl:text> `xml:&quot;,chardata&quot;`</xsl:text>
      <xsl:value-of select="$nl" />
      <xsl:value-of select="$nl" />
      <xsl:call-template name="simpleContent" />
    </xsl:for-each>
    <xsl:call-template name="complexContent" />
    <xsl:text>}</xsl:text>
    <xsl:value-of select="$nl" />
    <xsl:value-of select="$nl" />
  </xsl:template>

</xsl:stylesheet>
