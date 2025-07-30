# Natural Language Interface: Detailed Specification

## 1. Introduction

The Natural Language Interface (NLI) is the user-facing component of our AI-powered Internal Developer Platform. It serves as the primary interaction point between developers and the system, allowing users to express their infrastructure needs in conversational natural language. This component is designed to be intuitive, responsive, and context-aware, providing a seamless experience that abstracts away the complexity of underlying infrastructure management.

## 2. Component Overview

The NLI is responsible for:

1. **User Interaction**: Providing a conversational interface for developers to express their infrastructure needs
2. **Context Management**: Maintaining conversation history and user context across multiple interactions
3. **Input Validation**: Ensuring user input is clean and appropriate for processing
. **Output Presentation**: Displaying system responses in a clear, organized manner
5. **Error Handling**: Providing helpful feedback when system cannot understand or execute requests
6. **Progress Indication**: Showing status of long-running operations and tasks

## 3. Design Principles

### 3.1 Conversational Experience

The NLI adopts a chat-like interface that feels natural and intuitive to developers:

- **Multi-mode Communication**: Supports text, voice, and potential inputs where appropriate
- **Contextual Memory**: Remembers previous interactions within a session and across sessions
- **Prompt Templating**: Provides helpful suggestions and templates to guide user input
- **Interactive Clarification**: Asks follow-up questions when intent is unclear or parameters are missing

### 3.2 User Experience (UX) Design

The UX is designed to be clean, minimal, and functional:

1. **Input Area*:
    - Text input field with multi-line support
   - Voice input button (for supported browsers)
   - Suggested prompts and commands based on context
   - Attachment support for configuration files

2. **Conversation Area:
    - Chronological display of messages
    - User and system messages distinguished by color and iconography
   - Scrollable history with clear separation between sessions
   - Search and filter capabilities for history

3. **Status Indicators:
    - Typing indicator when system is processing
    - Read recipts wiun messages are delivered
    - System status notifications for long-running operations

4. **Help & Support:
    - Context-sensitive help button
    - Quick commands for common actions
    - Feedback and reporting mechanisms

5. **Settings & Preferences:*
    - User profile management
    - Language and tone preferences
    - Notification preferences
    - Theme and display customization options

### 3.3 Accessibility

The interface is designed with accessibility as a core principle:

- **Keyboard Navigation**: Full keyboard navigation and shortcuts
- **Screen Reader Support**: Compatibility with screen readers
- **High Contrast Mode**: Clear visual indicators and large text
- **Responsive Design**: Adapts to different screen sizes and orientations
- **K\Hİ\Ü
ŠˆÛÛ\]Xš[]HÚ]\ÜÚ\İ]™HXÚ›ÛÙÚY\Â‚ˆÈÈˆXÚšXØ[[\[Y[][Û‚‚ˆÈÈÈŒHœ›Û[™XÚ›ÛÙÚY\Â‚•Hœ›Û[™\ÈZ[\Ú[™È[Ù\›ˆÙXˆXÚ›ÛÙÚY\È›ÜˆX^[][HÛÛ\]Xš[]N‚‚‹H
Š”™XXİšœÊŠˆ›ÜˆHXZ[ˆ[\™˜XÙHÛÛ\Û™[Â‹H
Š•\TØÜš\
Šˆ›Üˆ\HØY™]H[™]™[Ü\ˆ^\šY[˜ÙB‹H
Š•ÙX”ÛØÚÙ]Êˆ›Üˆ™X[][YHÛÛ[][šXØ][ÛˆÚ]H˜XÚÙ[™‹H
ŠÜÜËZ[‹ZœÊˆ›Üˆİ[[™È[™[Z[™Â‹H
Š“X]\šX[URJˆÜˆ
Š•Z[Ú[™ÔÔÊŠˆ›ÜˆÛÛœÚ\İ[ÛÛ\Û™[\ÚYÛ‚‚ˆÈÈÈŒˆ˜XÚÙ[™\˜Ú]Xİ\™B‚•H˜XÚÙ[™Ù\šXÙ\ÈX[˜YÙ\ÈÛÛ™\œØ][Ûˆİ]K\Ù\ˆ]][XØ][Û‹[™[YÜ˜][ÛˆÚ]İ\ˆŞ\İ[HÛÛ\Û™[Î‚‚‹H
Š“›ÙKšœÊŠˆÜˆ
Ê”]ÛŠŠˆ›ÜˆHXZ[ˆ\XØ][ÛˆÙÚXÂ‹H
Š‘^™\ÜÊŠˆÜˆ
Š‘˜\İYš^JŠˆ›ÜˆHÙXˆ\XØ][Ûˆœ˜[Y]ÛÜšÂ‹H
Š”ÛØÚÙ]’SÊŠˆ›Üˆ™X[][YHÛÛ[][šXØ][Û‚‹H
Š”™Y\ÊŠˆÜˆ
Š“[Û™ÛÑŠŠˆ›ÜˆÛÛ™\œØ][Ûˆ\İÜHİÜ˜YÙB‹H
Š]]
ŠˆÙ\šXÙH›Üˆ\Ù\ˆY[]H[™\›Z\ÜÚ[ÛˆX[˜YÙ[Y[‚ˆÈÈÈŒÈT\È[™[™Ú[Â‚•H“H^ÜÙ\ÈÙ]™\˜[T\È›Üˆ[\›˜[ÛÛ\Û™[ÛÛ[][šXØ][Û‚‚ŒKˆ
ŠÛÛ™\œØ][ÛˆTJŠˆX[˜YÙ\ÈÛÛ™\œØ][Ûˆİ]KY\ÜØYÙ\Ë[™Ù\ÜÚ[ÛœÂˆHÔÕØ\KØÛÛ™\œØ][ÛœÈHİ\™]ÈÛÛ™\œØ][Û‚ˆHÑUØ\KØÛÛ™\œØ][ÛœËŞÚYHHÙ]ÛÛ™\œØ][Ûˆ]Z[ÂˆHUØ\KØÛÛ™\œØ][ÛœËŞÚYHH\]HÛÛ™\œØ][Û‚ˆHSUHØ\KØÛÛ™\œØ][ÛœËŞÚYHH[]HÛÛ™\œØ][Û‚‚‹ˆ
Š“Y\ÜØYÙHTJŠˆ[™\È[™]šYX[Y\ÜØYÙ\ÈÚ][ˆHÛÛ™\œØ][Û‚ˆHÔÕØ\KÛY\ÜØYÙ\ÈHÙ[™H™]ÈY\ÜØYÙBˆHÑUØ\KÛY\ÜØYÙ\ËŞÚYHHÙ]Y\ÜØYÙH]Z[Â‚ŒËˆ
Š•\Ù\ˆTJŠˆX[˜YÙ\È\Ù\ˆ›Ùš[\È[™™Y™\™[˜Ù\ÂˆHÑUØ\Kİ\Ù\‹Ü›Ùš[HHÙ]İ\œ™[\Ù\ˆ›Ùš[BˆHUØ\Kİ\Ù\‹Ü›Ùš[HH\]H\Ù\ˆ›Ùš[BˆHÑUØ\Kİ\Ù\‹Ü™Y™\™[˜Ù\ÈHÙ]\Ù\ˆ™Y™\™[˜Ù\Â‚ˆ
Š’[TJŠˆ›İšY\È[ÛÛ[[™İYÙÙ\İ[ÛœÂˆHÑUØ\KÚ[ØÛÛ^HÙ]ÛÛ^\Ù[œÚ]]™H[ˆHÑUØ\KÚ[ÜİYÙÙ\İ[ÛœÈHÙ]ÛÛ[X[™İYÙÙ\İ[ÛœÂ‚ˆÈÈÈ]H[Ù[Â‚ÛÛ™\œØ][ÛœÈ[™Y\ÜØYÙ\È\™HİÜ™Y\Ú[™ÈİXİ\™Y]H[Ù[Î‚‚˜Âˆ˜ÛÛ™\œØ][ÛˆˆÂˆšYˆ]ZY]˜[YH‹ˆ\Ù\—ÚYˆ\Ù\‹ZY‹ˆ]Hˆ“™]ÈTH›Üˆ\Ù\ˆX[˜YÙ[Y[‹ˆ˜Ü™X]YØ]ˆŒŒLKLMULŒŒˆ‹ˆ\]YØ]ˆŒŒLKLMULŒÌŒˆ‹ˆœİ]\Èˆ˜Xİ]™H‹ˆ›Y\ÜØYÙ\ÈˆÂˆÂˆšYˆ›Y\ÜØYÙKZYLH‹ˆ\Hˆ\Ù\ˆ‹ˆ˜ÛÛ[ˆ’H™YYH™]ÈTHÙ\™\ˆ‹ˆ[Y\İ[\ˆŒŒLKLMULŒŒˆ‹ˆ›Y]Y]HˆÂˆœÛİ\˜ÙHˆÙXˆ‹ˆ›[™İXYÙHˆ™[ˆ‚ˆBˆKˆÂˆšYˆ›Y\ÜØYÙKZYLˆ‹ˆ\HˆœŞ\İ[H‹ˆ˜ÛÛ[ˆ’HØ[ˆ[[İHÜ™X]HH™]ÈTHÙ\™\‹ˆÚ]Ûİ[[İHZÙHÈ˜[YH]È‹ˆ[Y\İ[\ˆŒŒLKLMULŒNLˆ‹ˆ›Y]Y]HˆÂˆ˜ÛÛ™šY[˜ÙHˆMKˆš[[ˆ˜Ü™X]WØ\WÜÙ\™\ˆ‚ˆBˆBˆBˆBŸB˜‚ˆÈÈKˆ[YÜ˜][ÛˆÚ]İ\ˆÛÛ\Û™[Â‚ˆÈÈÈKŒH[œ]Ûİ\˜Ù\Â‚•H“H™XÙZ]™\È[œ]œ›ÛN‚‹H
Š•\Ù\ˆ[\˜Xİ[ÛœÊŠˆ\™Xİ[œ]›İYÚHÙXˆ[\™˜XÙB‹H
Š•›ÚXÙH[œ]
ŠˆÜYXÚ]Ë]^ÛÛ™\œÚ[Ûˆ›Üˆ›ÚXÙHÛÛ[X[™Â‹H
‘š[H\ØYÊŠˆÛÛ™šYİ\˜][Ûˆš[\È[™Øİ[Y[][Û‚‹H
ÛÛ^X[[
ŠˆÛÛ^\Ù[œÚ]]™H[İYÙÙ\İ[ÛœÂ‚ˆÈÈÈKŒˆİ]]\İ[˜][ÛœÂ‚•H“HÙ[™È›ØÙ\ÜÙY[œ]Î‚‹H
Š’[[™XÛÙÛš][Ûˆ	ˆ\˜[Y]\ˆ^˜Xİ[ÛŠŠˆ›Üˆ[[[˜[\Ú\È[™\˜[Y]\ˆ^˜Xİ[Û‚‹H
Š]Y]	ˆ˜XÙXXš[]HŞ\İ[JŠˆ›ÜˆÙÙÚ[™È[™]Y]˜Z[Â‹H
Š‘Ûİ™\›˜[˜ÙH	ˆÛXŞH[™Ú[™JŠˆ›Üˆ\›Z\ÜÚ[ÛˆÚXÚÜÈ[™˜[Y][Û‚‹H
Š“PÔP˜\ÙYÛÛŞ\İ[JŠˆ›Üˆ^Xİ][™È[™œ˜\İXİ\™HÜ\˜][ÛœÂ‚ˆÈÈÈKŒÈ]™[›İÂ‚•H“H\XÚ\]\È[ˆ[ˆ]™[Yš]™[ˆ\˜Ú]Xİ\™N‚‚˜\Ù\—Ú[œ]Oˆ“HOˆ[[Ü™XÛÙÛš][Ûˆ™\]Y\İ››H™\ÜÛœÙHH[[™XÛÙÛš][Ûˆ	ˆ\˜[Y]\ˆ^˜Xİ[Û‚œŞ\İ[WØXİ[ÛˆHPÔÛÛŞ\İ[Bœİ]\×İ\]HOˆ›İYšXØ][ÛˆÙ\šXÙB˜‚ˆÈÈ‹ˆÙXİ\š]H[™š]˜XŞB‚ˆÈÈÈ‹ŒH]][XØ][Û‚‚•H“H[\[Y[È›Ø\İ]][XØ][ÛˆYXÚ[š\Û\Î‚‚‹H
Š”Ù\ÜÚ[ÛˆX[˜YÙ[Y[
ŠˆÙXİ\™HÙ\ÜÚ[ÛˆÜ™X][Ûˆ[™X[˜YÙ[Y[‹H
Š’•ÕÚÙ[œÊŠˆİ][\ÜÈ]][XØ][ÛˆÚÙ[œÈ›ÜˆTHXØÙ\ÜÂ‹H
Š”Ú[™ÛHÚYÛ‹SÛˆ
ÔÓÊJŠˆ[YÜ˜][ÛˆÚ]Ü™Ø[š^˜][Û˜[Y[]H›İšY\œÂ‹H
Š“][KQ˜XİÜˆ]][XØ][ÛŠŠˆİ\Ü›ÜˆY][Û˜[]][XØ][Ûˆ˜XİÜœÂ‚ˆÈÈÈ‹Œˆ\›Z\ÜÚ[ÛœÂ‚•\Ù\ˆ\›Z\ÜÚ[ÛœÈ\™H[™›Ü˜ÙY]][\H]™[Î‚‚‹H
Š”›ÛKP˜\ÙYXØÙ\ÜÊŠˆY™™\™[\›Z\ÜÚ[ÛœÈ›Üˆ]™[Ü\œËÜ\˜]ÜœË[™YZ[š\İ˜]ÜœÂ‹H
Š”™\Ûİ\˜ÙH][İ\ÊŠˆ[Z]ÈÛˆH[X™\ˆ[™\HÙˆ™\Ûİ\˜Ù\ÈH\Ù\ˆØ[ˆÜ™X]B‹H
ŠÛÜİÛÛ›ÛÊŠˆ™]™[[ÛˆÙˆ^[œÚ]™H[™œ˜\İXİ\™HÜ™X][Û‚‹H
Š\›İ˜[ÛÜšÙ›İÜÊŠˆ™\]Z\™[Y[›Üˆ\›İ˜[›ÜˆYÚZ[\XİÜ\˜][ÛœÂ‚ˆÈÈÈ‹ŒÈ]H›İXİ[Û‚‚•HŞ\İ[H›İXİÈ\Ù\ˆ]H[™ÛÛ™\œØ][ÛœÎ‚‚‹H
Š‘[˜Ü\[ÛŠŠˆ[ÛÛ™\œØ][ÛœÈ\™H[˜Ü\Y]™\İ‹H
Š‘]H™][[ÛŠŠˆÛÛ™\œØ][ÛœÈ\™H™]Z[™Y›ÜˆHYš[™Y\š[Ù‹H
Š[›Û[Z^˜][ÛŠŠˆ\Ù\ˆY[]Y\È\™H›İXİY[™›İ[šÙYÈ\œÛÛ˜[]B‹H
ŠÛÛ\X[˜ÙJŠˆ™Yİ[\ˆ]Y]ÈÈ[œİ\™HÛÛ\X[˜ÙHÚ]]H›İXİ[Ûˆ™Yİ[][ÛœÂ‚ˆÈÈÈ‹ÙXİ\š]H[Ûš]Üš[™Â‚•HŞ\İ[H[˜ÛY\ÈÛÛ\™Z[œÚ]™HÙXİ\š]H[Ûš]Üš[™Î‚‚‹H
Š]Y]ÙÙÚ[™ÊŠˆ[\Ù\ˆXİ[ÛœÈ[™Ş\İ[H™\ÜÛœÙ\È\™HÙÙÙY‹H
’[\Ú[Ûˆ]Xİ[ÛŠˆ[Ûš]Üš[™È›Üˆİ\ÜXÚ[İ\ÈXİ]š]H[™İ[X[ÙXİ\š]H™X]Â‹H
”˜]H[Z][™Êˆ›İXİ[ÛˆYØZ[œİœ]H›Ü˜ÙH]XÚÜÂ‹H
Š•[™\˜Xš[]HØØ[›š[™Ê
ˆ™Yİ[\ˆÙXİ\š]HØØ[œÈÈY[YH[™\˜Xš[]Y\Â‚ˆÈÈËˆ\™›Ü›X[˜ÙHÜ[Z^˜][Û‚‚ˆÈÈÈËŒHœ›Û[™Ü[Z^˜][Û‚‚•Hœ›Û[™\ÈÜ[Z^™Y›Üˆ˜\İØY[™È[™™\ÜÛœÚ]™[™\ÜÎ‚‚‹H
ŠÛÙHÜ][™Ê
ˆ[™\È\™HÜ][™^K[ØYYÛˆ[X[™‹H
ŠØXÚ[™ÊŠˆİ]XÈ\ÜÙ]È\™HØXÚY[™Ù\™Yœ›ÛHÑ‚‹H
Š’[XYÙHÜ[Z^˜][ÛŠŠˆ[XYÙ\È\™HÛÛ\™\ÜÙY[™Ü[Z^™Y›ÜˆÙXˆ[]™\B‹H
Š“Z[šYšXØ][ÛŠŠˆ[›™XÙ\ÜØ\HÛÙH\È™[[İ™Y[™\ÜÙ]È\™HZ[šYšYY‚ˆÈÈÈËŒˆ˜XÚÙ[™Ü[Z^˜][Û‚‚•H˜XÚÙ[™\ÈÜ[Z^™Y›ÜˆØØ[Xš[]H[™\™›Ü›X[˜ÙN‚‚‹H
ŠÛÛ›™Xİ[ÛˆÛÛ[™ÊŠˆ]X˜\ÙHÛÛ›™Xİ[ÛœÈ\™HÛÛY[™™]\ÙY‹H
Š”™Y\ÈØXÚ[™Ê
ˆœ™\]Y[HXØÙ\ÜÙY]H\ÈØXÚY[ˆY[[ÜB‹H
Š\Ş[˜Ú›Û›İ\È›ØÙ\ÜÚ[™ÊŠˆÛ™Ë\[›š[™È\ÚÜÈ\™H[™Y\Ş[˜Ú›Û›İ\ÛB‹H
Š“ØY˜[[˜Ú[™Ê
ˆ™\]Y\İÈ\™H\İšX]YXÜ›ÜÜÈ][\H[œİ[˜Ù\È›ÜˆYÚ]˜Z[Xš[]B‚ˆÈÈÈËŒÈ[Ûš]Üš[™È[™[\[™Â‚•HŞ\İ[H[˜ÛY\ÈÛÛ\™Z[œÚ]™H[Ûš]Üš[™Î‚‹H
Š”\™›Ü›X[˜ÙHY]šXÜÊŠˆ˜XÚÈ™\ÜÛœÙH[Y\Ë\œ›Üˆ˜]\Ë[™\ØYÙH]\›œÂ‹H
Š’X[ÚXÚÜÊŠˆ[Ûš]ÜˆŞ\İ[HX[[™\™›Ü›X[˜ÙB‹H
Š[\[™Ê
ˆÛÛ™šYİ\˜X›H[\È›ÜˆÙXİ\š]H]™[Ë\™›Ü›X[˜ÙH\ÜİY\Ë[™Ş\İ[H\œ›ÜœÂ‹H
“ÙÈ[˜[\Ú\ÊŠˆ]]ÛX]Y[˜[\Ú\ÈÙˆÙÜÈÈY[YH\ÜİY\È[™Ü[Z^˜][ÛˆÜÜ[š]Y\Â‚ˆÈÈˆ\İ[™Èİ˜]YŞB‚ˆÈÈÈŒH[š]\İ[™Â‚ÛÛ\Û™[È\™H\İY[ˆ\ÛÛ][ÛˆÈ[œİ\™H™[XXš[]N‚‹H
Š‘œ›Û[™ÛÛ\Û™[\İÊŠˆ[š]\İÈ›ÜˆRHÛÛ\Û™[Ë[\˜Xİ[ÛˆÙÚXË[™İ]HX[˜YÙ[Y[‹H
Š˜XÚÙ[™TH\İÊŠˆ\İÈ›ÜˆTH[™Ú[Ë]H˜[Y][Û‹[™\Ú[™\ÜÈÙÚXÂ‹H
Š’[YÜ˜][Ûˆ\İÊŠˆ\İÈ›Üˆ[YÜ˜][ÛˆÚ]İ\ˆŞ\İ[HÛÛ\Û™[Â‚ˆÈÈÈŒˆ[YÜ˜][Ûˆ\İ[™Â‚•HŞ\İ[H\È\İY[™]ËY[™È[œİ\™HÙX[[\ÜÈ[YÜ˜][Û‚‹H
Š‘[™]ËQ[™\İİZ]\ÊŠˆÛÛ\]H\Ù\ˆ›İ\›™^\Èœ›ÛH[œ]È[™œ˜\İXİ\™H\Ş[Y[‹H
Š”\™›Ü›X[˜ÙH\İ[™ÊŠˆØY\İ[™ÈÈ[œİ\™HŞ\İ[HØ[ˆ[™H^XİY\Ù\ˆØY‹H
Š”ÙXİ\š]H\İ[™ÊŠˆ[™]˜][Ûˆ\İ[™ÈÈY[YH[™\˜Xš[]Y\È[™ÙXİ\š]H›]ÜÂ‹H
ŠXØÙ\[˜ÙH\İ[™ÊŠˆ\Ù\ˆXØÙ\[˜ÙH\İ[™ÈÈ[œİ\™HH[\™˜XÙH\È[Z]]™H[™X\ŞHÈ\ÙB‚ˆÈÈÈŒÈ\Ù\ˆXØÙ\[˜ÙH\İ[™Â‚™Y›Ü™H\Ş[Y[HŞ\İ[H[™\™ÛÙ\È\Ù\ˆXØÙ\[˜ÙH\İ[™Î‚‹H
Š[H\İ[™ÊŠˆ™[X\ÙHÈHÛX[Ü›İ\Ùˆ\Ù\œÈ›Üˆ™YY˜XÚÈ[™YÈš^[™Â‹H
Š™]H\İ[™ÊŠˆÛÛ[YY\İ[™ÈÚ][˜Ü™X\Ú[™ÛH\™Ù\ˆÜ›İ\ÈÙˆ\Ù\œÂ‹H
‘™YY˜XÚÈÛÛXİ[ÛŠˆ™Yİ[\ˆÛÛXİ[ÛˆÙˆ\Ù\ˆ™YY˜XÚÈ[™İYÙÙ\İ[ÛœÂ‹H
•V™\ÙX\˜Ú
ˆÛÛ[[İ\È[\›İ™[Y[˜\ÙYÛˆ\Ù\ˆ™\ÙX\˜Ú[™™Z]š[Ü˜[[˜[\Ú\Â‚ˆÈÈKˆ\Ş[Y[[™Ü\˜][ÛœÂ‚ˆÈÈÈKŒH\Ş[Y[İ˜]YŞB‚•H“H\È\ÚYÛ™Y›ÜˆÛÛZ[™\š^™Y\Ş[Y[\Ú[™È[Ù\›ˆ]“ÜÈ˜XİXÙ\Î‚‹H
ŠÛÛZ[™\ˆÜ˜Ú\İ˜][ÛŠŠˆØÚÙ\ˆÛÛZ[™\œÈ›ÜˆX[H[š\›Û›Y[\Ş[Y[‹H
Š’İX™\›™]\È\Ş[Y[
ŠˆİX™\›™]\ÈX[šY™\İÈ[™[HÚ\È›ÜˆX[˜YÙ[Y[‹H
ŠÛÛ™šYİ\˜][Ûˆ\ÈÛÙJŠˆ[ÛÛ™šYİ\˜][Ûˆ\È™\œÚ[Û‹XÛÛ›ÛY[™\ŞYYÚ]H\XØ][Û‚‹H
Š‘[š\›Û›Y[˜\šXX›\ÊŠˆY™™\™[ÛÛ™šYİ\˜][ÛœÈ›Üˆ]™[ÜY[İYÚ[™Ë[™›ÙXİ[Ûˆ[š\›Û›Y[Â‚ˆÈÈÈKŒˆÛ™ÛÚ[™ÈÜ\˜][ÛœÂ‚•HŞ\İ[H[˜ÛY\ÈÛÛ\™Z[œÚ]™HÛ™ÛÚ[™ÈÜ\˜][ÛœÈ[™XZ[[˜[˜ÙNˆ‹H
’X[ÚXÚÜÊˆ[Ûš]Üš[™ÈÙˆÛÛ\Û™[X[[™\™›Ü›X[˜ÙB‹H
“ÙÈX[˜YÙ[Y[
ˆÙ[˜[^™YÙÙÈÛÛXİ[Ûˆ[™[˜[\Ú\Â‹H
“Y]šXÜÈÛÛXİ[ÛŠˆÛÛXİ[ÛˆÙˆ\ØYÙH[™\™›Ü›X[˜ÙHY]šXÜÂ‹H˜XÚİ\[™™XÛİ™\Jˆˆ™Yİ[\ˆ˜XÚİ\È[™\Ø\İ\ˆ™XÛİ™\H›ØÙY\™\Â‹H
•\]\È[™]Ú[™Ê
ˆÛÛ›ÛY\]H›ØÙ\ÜÈÚ]Z[š[X[İÛ[YB‚ˆÈÈÈKŒÈØØ[[™Èİ˜]YŞB‚•HŞ\İ[H\È\ÚYÛ™YÈØØ[HÜš^›Û[H[™™\XØ[N‚‹H
’Üš^›Û[ØØ[[™ÊˆY[™È[Ü™H[œİ[˜Ù\ÈÈ[™H[˜Ü™X\ÙY\Ù\ˆØY‹H
•™\XØ[ØØ[[™ÊˆY[™È[Ü™H^Y\œÈÙˆ[™œ˜\İXİ\™HÈ[™H[˜Ü™X\ÙYÛÛ\^]B‹H
‘]X˜\ÙHØØ[[™ÊˆY[™È[Ü™H]X˜\ÙH[œİ[˜Ù\ÈÈ[™H[˜Ü™X\ÙY]H›Û[YB‹H
ØXÚ[™ÈØØ[[™ÊˆY[™È[Ü™HØXÚ[™È^Y\œÈÈ[\›İ™H\™›Ü›X[˜ÙH[™\ˆX]HØY‚ˆÈÈLˆÛÛ˜Û\Ú[Û‚‚•H˜]\˜[[™İXYÙH[\™˜XÙHÙ\™\È\ÈHÜš]XØ[[HÚ[›Üˆ]™[Ü\œÈ[Èİ\ˆRK\İÙ\™Y[\›˜[]™[Ü\ˆ]›Ü›KˆH›İšY[™È[ˆ[Z]]™KÛÛ^X]Ø\™K[™™\ÜÛœÚ]™H[\™˜XÙK]Xœİ˜XİÈ]Ø^HHÛÛ\^]HÙˆ[™œ˜\İXİ\™H›İš\Ú[Ûš[™ÈÚ[HXZ[Z[š[™ÈHÛİ™\›˜[˜ÙH[™ÛÛœÚ\İ[˜ŞH™\]Z\™Y›Üˆ[\œš\ÙH\Ş[Y[‚‚•›İYÚØ\™Y[\ÚYÛˆÙˆH\Ù\ˆ^\šY[˜ÙK›Ø\İXÚšXØ[[\[Y[][Û‹[™ÛÛ\™Z[œÚ]™HÙXİ\š]HYX\İ\™\ËH“H[œİ\™\È]]™[Ü\œÈØ[ˆY™™Xİ]™[HÛÛ[][šXØ]HZ\ˆ[™œ˜\İXİ\™H™YYÈÚ[H›İšY[™ÈHŞ\İ[HÚ]HİXİ\™Y[œ]™\]Z\™Y›Üˆ™[XX›H[™Ûİ™\›™Y^Xİ][Û‹‚‚•\ÈÛÛ\Û™[[X›ÙY\Èİ\ˆš[˜Ú\HÙˆXZÚ[™È[™œ˜\İXİ\™HX[˜YÙ[Y[XØÙ\ÜÚX›HÚ[HXZ[Z[š[™ÈHİ[™\™È[™ÛÛ›ÛÈ™\]Z\™Y›Üˆ[\œš\ÙHÜ\˜][ÛœËˆ]Ù]ÈHİYÙH›ÜˆH[\™HŞ\İ[HÈ˜[œÙ›Ü›H˜]\˜[[™İXYÙH[ÈXİ[Û˜X›KÛİ™\›™Y[™œ˜\İXİ\™