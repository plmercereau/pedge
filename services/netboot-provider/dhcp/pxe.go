package dhcp

import (
    "encoding/binary"
	"fmt"
)

// ArchType represents different PXE client system architecture types.
type ArchType uint16

// Helper function to convert byte slice to ArchType
func BytesToArchType(data []byte) (ArchType, error) {
    if len(data) < 2 {
        return 0, fmt.Errorf("invalid data length: %d", len(data))
    }
    return ArchType(binary.BigEndian.Uint16(data)), nil
}

const (
    ArchTypeIntelx86PC ArchType = iota
    ArchTypeNECPC98
    ArchTypeEFIItanium
    ArchTypeDECAlpha
    ArchTypeArcx86
    ArchTypeIntelLeanClient
    ArchTypeEFIIA32
    ArchTypeEFI64
    ArchTypeEFIXscale
    ArchTypeEFIX8664

    ArchTypeEFIBC = 10 + iota - 10
    ArchTypeEFIX8664v2
    ArchTypeMicrosoftXbox
    ArchTypeEFIARM
    ArchTypeEFIARM32Bit
    ArchTypeEFIARM64Bit
    ArchTypeEFIARM32BitSecureBoot
    ArchTypeEFIARM64BitSecureBoot
    ArchTypeEFIx86NetworkBoot
    ArchTypeEFIX8664NetworkBoot
    ArchTypeEFIHTTPBootIA32
    ArchTypeEFIHTTPBootX8664
    ArchTypeEFIHTTPBootARM32
    ArchTypeEFIHTTPBootARM64
    ArchTypeEFIx86BIOS
    ArchTypeEFIx8664BIOS
    ArchTypeEFIHTTPBootIA32Authenticated
    ArchTypeEFIHTTPBootX8664Authenticated
    ArchTypeEFIHTTPBootARM32Authenticated
    ArchTypeEFIHTTPBootARM64Authenticated
)

// String returns the human-readable name of the architecture type.
func (at ArchType) String() string {
    names := map[ArchType]string{
        ArchTypeIntelx86PC:                  "Intel x86PC",
        ArchTypeNECPC98:                     "NEC/PC98",
        ArchTypeEFIItanium:                  "EFI Itanium",
        ArchTypeDECAlpha:                    "DEC Alpha",
        ArchTypeArcx86:                      "Arc x86",
        ArchTypeIntelLeanClient:             "Intel Lean Client",
        ArchTypeEFIIA32:                     "EFI IA32",
        ArchTypeEFI64:                       "EFI BC",
        ArchTypeEFIXscale:                   "EFI Xscale",
        ArchTypeEFIX8664:                    "EFI x86-64",
        ArchTypeEFIBC:                       "EFI BC",
        ArchTypeEFIX8664v2:                  "EFI x86-64 v2",
        ArchTypeMicrosoftXbox:               "Microsoft Xbox",
        ArchTypeEFIARM:                      "EFI ARM",
        ArchTypeEFIARM32Bit:                 "EFI ARM 32-bit",
        ArchTypeEFIARM64Bit:                 "EFI ARM 64-bit",
        ArchTypeEFIARM32BitSecureBoot:       "EFI ARM 32-bit (Secure Boot)",
        ArchTypeEFIARM64BitSecureBoot:       "EFI ARM 64-bit (Secure Boot)",
        ArchTypeEFIx86NetworkBoot:           "EFI x86 (Network Boot)",
        ArchTypeEFIX8664NetworkBoot:         "EFI x86-64 (Network Boot)",
        ArchTypeEFIHTTPBootIA32:             "EFI IA32 (UEFI HTTP Boot)",
        ArchTypeEFIHTTPBootX8664:            "EFI x86-64 (UEFI HTTP Boot)",
        ArchTypeEFIHTTPBootARM32:            "EFI ARM 32-bit (UEFI HTTP Boot)",
        ArchTypeEFIHTTPBootARM64:            "EFI ARM 64-bit (UEFI HTTP Boot)",
        ArchTypeEFIx86BIOS:                  "EFI x86 BIOS",
        ArchTypeEFIx8664BIOS:                "EFI x86-64 BIOS",
        ArchTypeEFIHTTPBootIA32Authenticated:"EFI IA32 (UEFI HTTP Boot, Authenticated)",
        ArchTypeEFIHTTPBootX8664Authenticated:"EFI x86-64 (UEFI HTTP Boot, Authenticated)",
        ArchTypeEFIHTTPBootARM32Authenticated:"EFI ARM 32-bit (UEFI HTTP Boot, Authenticated)",
        ArchTypeEFIHTTPBootARM64Authenticated:"EFI ARM 64-bit (UEFI HTTP Boot, Authenticated)",
    }
    return names[at]
}
