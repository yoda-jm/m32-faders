$(document).ready(function() {
    let allFaders = [];
    let currentFaderId = null;

    // Initial UI State
    $('#fader-slider').prop('disabled', true);
    $('#mute-button').prop('disabled', true);
    $('#fader-item-select').prop('disabled', true);

    // Helper function to update fader UI elements
    function updateFaderUI(faderData) {
        if (!faderData) {
            $('#selected-fader-name').text('No Fader Selected');
            $('#fader-slider').val(0).prop('disabled', true);
            $('#fader-db-value').text('0.0 dB');
            $('#mute-button').text('Mute').removeClass('btn-success').addClass('btn-danger').prop('disabled', true);
            currentFaderId = null;
            return;
        }

        currentFaderId = faderData.ID;
        $('#selected-fader-name').text(`${faderData.Name} (${faderData.ID})`);
        $('#fader-slider').val(faderData.Level).prop('disabled', false);
        $('#fader-db-value').text(`${parseFloat(faderData.Level).toFixed(1)} dB`);

        if (faderData.Muted) {
            $('#mute-button').text('Unmute').removeClass('btn-danger').addClass('btn-success').prop('disabled', false);
        } else {
            $('#mute-button').text('Mute').removeClass('btn-success').addClass('btn-danger').prop('disabled', false);
        }
    }

    // 1. Fetch All Faders and Populate Type Selector
    $.ajax({
        url: '/api/faders',
        method: 'GET',
        dataType: 'json',
        success: function(data) {
            allFaders = data;
            let faderTypes = ['All']; // Add 'All' as an option
            let types = new Set(allFaders.map(fader => fader.Type));
            types.forEach(type => faderTypes.push(type));
            
            $('#fader-type-select').empty().append($('<option>', {
                selected: true,
                disabled: true,
                text: 'Select Type...'
            }));

            faderTypes.forEach(function(type) {
                $('#fader-type-select').append($('<option>', {
                    value: type,
                    text: type
                }));
            });
        },
        error: function(jqXHR, textStatus, errorThrown) {
            console.error('Error fetching faders:', textStatus, errorThrown);
            // Potentially display an error to the user
        }
    });

    // 2. Populate Fader Item Selector (On Type Change)
    $('#fader-type-select').on('change', function() {
        const selectedType = $(this).val();
        $('#fader-item-select').empty().append($('<option>', {
            selected: true,
            disabled: true,
            text: 'Select Item...'
        })).prop('disabled', true);
        updateFaderUI(null); // Reset fader display

        if (!selectedType || selectedType === "Select Type...") {
            return;
        }

        const filteredFaders = (selectedType === 'All') 
            ? allFaders 
            : allFaders.filter(fader => fader.Type === selectedType);

        if (filteredFaders.length > 0) {
            filteredFaders.sort((a, b) => a.ID.localeCompare(b.ID)); // Sort for consistent order
            filteredFaders.forEach(function(fader) {
                $('#fader-item-select').append($('<option>', {
                    value: fader.ID,
                    text: `${fader.Name} (${fader.ID})` 
                }));
            });
            $('#fader-item-select').prop('disabled', false);
        }
    });

    // 3. Fetch and Display Fader Details (On Item Change)
    $('#fader-item-select').on('change', function() {
        const faderId = $(this).val();
        if (!faderId || faderId === "Select Item...") {
            updateFaderUI(null);
            return;
        }

        $.ajax({
            url: `/api/faders/${faderId}`,
            method: 'GET',
            dataType: 'json',
            success: function(faderData) {
                updateFaderUI(faderData);
            },
            error: function(jqXHR, textStatus, errorThrown) {
                console.error(`Error fetching fader ${faderId}:`, textStatus, errorThrown);
                updateFaderUI(null);
                // Potentially display an error message to the user
            }
        });
    });

    // 4. Update Fader Level (On Slider Input)
    $('#fader-slider').on('input', function() {
        const newLevel = parseFloat($(this).val());
        $('#fader-db-value').text(`${newLevel.toFixed(1)} dB`);

        if (currentFaderId) {
            // Debounce or throttle this if performance becomes an issue
            $.ajax({
                url: `/api/faders/${currentFaderId}`,
                method: 'POST',
                contentType: 'application/json',
                data: JSON.stringify({ level: newLevel }),
                dataType: 'json',
                success: function(updatedFader) {
                    // console.log(`Fader ${currentFaderId} level updated to ${updatedFader.Level}`);
                    // Optionally, re-update UI strictly from server response if needed
                    // updateFaderUI(updatedFader); // Can cause slider jitter if not careful
                },
                error: function(jqXHR, textStatus, errorThrown) {
                    console.error(`Error updating fader ${currentFaderId} level:`, textStatus, errorThrown);
                    // Potentially revert slider or notify user
                }
            });
        }
    });

    // 5. Toggle Mute (On Mute Button Click)
    $('#mute-button').on('click', function() {
        if (!currentFaderId) return;

        const isCurrentlyMuted = $(this).hasClass('btn-success'); // If it has btn-success, it means it's showing "Unmute"
        const newMuteState = !isCurrentlyMuted;

        $.ajax({
            url: `/api/faders/${currentFaderId}`,
            method: 'POST',
            contentType: 'application/json',
            data: JSON.stringify({ muted: newMuteState }),
            dataType: 'json',
            success: function(updatedFader) {
                // console.log(`Fader ${currentFaderId} mute state updated to ${updatedFader.Muted}`);
                updateFaderUI(updatedFader); // Update UI with server response
            },
            error: function(jqXHR, textStatus, errorThrown) {
                console.error(`Error updating fader ${currentFaderId} mute state:`, textStatus, errorThrown);
                // Potentially notify user
            }
        });
    });
});
